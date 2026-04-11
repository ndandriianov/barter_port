import { useEffect, useMemo } from "react";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";

interface UseDraftOfferCountsOptions {
  enabled?: boolean;
}

function useDraftOfferCounts({ enabled = true }: UseDraftOfferCountsOptions = {}) {
  const dispatch = useAppDispatch();
  const {
    data: draftRefs,
    isLoading,
    error,
  } = dealsApi.useGetMyDraftDealsQuery(undefined, {
    skip: !enabled,
  });

  const draftIds = useMemo(
    () => draftRefs?.map((draft) => draft.id) ?? [],
    [draftRefs],
  );

  useEffect(() => {
    if (!enabled || draftIds.length === 0) {
      return;
    }

    const subscriptions = draftIds.map((draftId) =>
      dispatch(dealsApi.endpoints.getDraftDealById.initiate(draftId)),
    );

    return () => {
      subscriptions.forEach((subscription) => subscription.unsubscribe());
    };
  }, [dispatch, draftIds, enabled]);

  const drafts = useAppSelector((state) =>
    draftIds
      .map((draftId) => dealsApi.endpoints.getDraftDealById.select(draftId)(state).data)
      .filter((draft): draft is NonNullable<typeof draft> => Boolean(draft)),
  );

  const countsByOfferId = useMemo(() => {
    const counts: Record<string, number> = {};

    drafts.forEach((draft) => {
      const offerIds = new Set(draft.offers.map((offer) => offer.id));
      offerIds.forEach((offerId) => {
        counts[offerId] = (counts[offerId] ?? 0) + 1;
      });
    });

    return counts;
  }, [drafts]);

  return {
    countsByOfferId,
    isLoading,
    error,
  };
}

export default useDraftOfferCounts;
