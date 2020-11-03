package actorstate

import (
	"context"

	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/types"
	sa0builtin "github.com/filecoin-project/specs-actors/actors/builtin"
	sa2builtin "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"go.opentelemetry.io/otel/api/global"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/sentinel-visor/metrics"
	"github.com/filecoin-project/sentinel-visor/model"
)

type MultiSigActorExtractor struct {
	PrevState multisig.State

	CurrActor *types.Actor
	CurrState multisig.State
	CurrTs    *types.TipSet
}

func init() {
	Register(sa0builtin.MultisigActorCodeID, MultiSigActorExtractor{})
	Register(sa2builtin.MultisigActorCodeID, MultiSigActorExtractor{})
}

func (m MultiSigActorExtractor) Extract(ctx context.Context, a ActorInfo, node ActorStateAPI) (model.Persistable, error) {
	ctx, span := global.Tracer("").Start(ctx, "MultiSigActor")
	defer span.End()

	stop := metrics.Timer(ctx, metrics.ProcessingDuration)
	defer stop()

}

func NewMultiSigExtractionContext(ctx context.Context, a ActorInfo, node ActorStateAPI) (*MultiSigActorExtractor, error) {
	curTipset, err := node.ChainGetTipSet(ctx, a.TipSet)
	if err != nil {
		return nil, xerrors.Errorf("loading current tipset %s: %w", a.TipSet.String(), err)
	}

	curState, err := multisig.Load(node.Store(), &a.Actor)
	if err != nil {
		return nil, xerrors.Errorf("loading current multisig state at head %s: %w", a.Actor.Head, err)
	}

	prevState := curState
	if a.Epoch != 0 {
		prevActor, err := node.StateGetActor(ctx, a.Address, a.ParentTipSet)
		if err != nil {
			return nil, xerrors.Errorf("loading previous multisig %s at tipset %s epoch %d: %w", a.Address, a.ParentTipSet, a.Epoch)
		}

		prevState, err = multisig.Load(node.Store(), prevActor)
		if err != nil {
			return nil, xerrors.Errorf("loading previous multisig actor state: %w", err)
		}
	}

	return &MultiSigActorExtractor{
		PrevState: prevState,
		CurrActor: &a.Actor,
		CurrState: curState,
		CurrTs:    curTipset,
	}, nil
}
