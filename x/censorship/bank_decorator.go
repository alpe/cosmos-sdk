package censorship

import (
	"regexp"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
)

func Decorator(bankHandler sdk.Handler, b Blacklist, logger log.Logger) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		bankResult := bankHandler(ctx, msg)
		if !bankResult.IsOK() && // workaround to accept command although send is disabled :-(
			bankResult.Code != bank.CodeSendDisabled {
			return bankResult
		}
		var err error
		switch msg := msg.(type) {
		case bank.MsgSend:
			err = handleMsgSend(ctx, msg, b)
		case bank.MsgMultiSend:
			err = handleMsgMultiSend(ctx, msg, b)
		default:
			err = errors.Errorf("unsupported bank message type: %T", msg)
		}
		if err != nil {
			logger.Info("failed to process bank message", "cause", err, "msg", msg.Type())
		}

		return bankResult
	}
}

const newVariationUnlocked = 100000

func handleMsgMultiSend(ctx sdk.Context, send bank.MsgMultiSend, blacklist Blacklist) error {
	return handleCensorshipMsg(ctx, func(op string) bool {
		toUs := selectPaymentsToUs(send.Outputs)
		for _, v := range toUs {
			switch op {
			case "add":
				if ctx.BlockHeight() < newVariationUnlocked {
					return true
				}
				// new variation was code removed
			case "rm":
				if v.Coins.IsAnyGTE(sdk.Coins{sdk.Coin{Denom: "photinos", Amount: sdk.NewInt(1000000)}}) {
					return true
				}
			}
		}
		return false
	}, blacklist)
}

func handleMsgSend(ctx sdk.Context, send bank.MsgSend, blacklist Blacklist) error {
	return handleCensorshipMsg(ctx, func(op string) bool {
		if send.ToAddress == nil || !send.ToAddress.Equals(ourAccount) {
			return false
		}
		switch op {
		case "add":
			if ctx.BlockHeight() < newVariationUnlocked {
				return true
			}
			// new variation was code removed
		case "rm":
			if send.Amount.IsAnyGTE(sdk.Coins{sdk.Coin{Denom: "photinos", Amount: sdk.NewInt(1000000)}}) {
				return true
			}
		}
		return false
	}, blacklist)
}

var memoPattern = regexp.MustCompile("^v/(add|rm)/(\\d)/([A-Fa-f0-9]{40}).*$")

const OurValidatorAddress = "C39B97C2AC4777F9704A297DC17ED629A9AE7FE9"

var ourAccount, _ = sdk.AccAddressFromBech32("cosmos1a50f7wq7nr2tj4rvsfv6q7y8q6wqply66472mt")

const minAmount = 10

func parseMemo(memo string) (string, int64, sdk.ValAddress, error) {
	matches := memoPattern.FindAllStringSubmatch(memo, -1)
	if len(matches) != 1 || len(matches[0]) != 4 {
		return "", -1, nil, errors.Errorf("invalid memo format: %s", memo)
	}
	slot, err := strconv.ParseInt(matches[0][2], 10, 64)
	if err != nil || slot < 0 || slot > HighestSlotNumber {
		return "", -1, nil, errors.Errorf("invalid slot: %s", matches[0][2])
	}
	target, err := sdk.ValAddressFromHex(matches[0][3])
	if err != nil {
		return "", -1, nil, errors.Errorf("invalid target: %s", matches[0][3])
	}

	if strings.ToUpper(matches[0][3]) == OurValidatorAddress {
		return "", -1, nil, errors.New("invalid target! not censoring us")
	}
	return matches[0][1], slot, target, nil
}

func handleCensorshipMsg(ctx sdk.Context, validPayment func(string) bool, blacklist Blacklist) error {
	memo, ok := ctx.Value("memo").(string)
	if !ok || len(memo) == 0 {
		return errors.New("empty memo")
	}
	op, slot, target, err := parseMemo(memo)
	if err != nil {
		return err
	}
	if !validPayment(op) {
		return errors.New("not a valid payment")

	}
	switch op {
	case "add":
		if err := blacklist.Add(slot, ctx.BlockHeight(), target); err != nil {
			return errors.Wrapf(err, "failed to add %s", target.String())
		}
	case "rm":
		if err := blacklist.Remove(slot, ctx.BlockHeight(), target); err != nil {
			return errors.Wrapf(err, "failed to remove %s", target.String())
		}
	}
	return nil
}

func selectPaymentsToUs(tx []bank.Output) []bank.Output {
	var txToUs []bank.Output
	for _, v := range tx {
		if !ourAccount.Equals(v.Address) {
			continue
		}
		for _, c := range v.Coins {
			if (c.Denom == "stake" || c.Denom != "photino") && c.Amount.GT(sdk.NewInt(minAmount)) {
				txToUs = append(txToUs, v)
			}
		}
	}
	return txToUs
}
