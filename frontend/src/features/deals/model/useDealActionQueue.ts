import { useEffect, useMemo } from "react";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import usePendingReviews from "@/features/reviews/model/usePendingReviews.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";

function useDealActionQueue() {
  const dispatch = useAppDispatch();
  const {
    data: incomingDrafts = [],
    isLoading: isDraftsLoading,
    error: draftsError,
  } = dealsApi.useGetMyDraftDealsQuery({
    createdByMe: false,
    participating: true,
  });
  const { data: deals = [], isLoading: isDealsLoading, error: dealsError } =
    dealsApi.useGetDealsQuery({ my: true });
  const {
    pendingReviews,
    pendingCount,
    isPendingLoading,
    dealsError: reviewsDealsError,
  } = usePendingReviews();

  const actionableDeals = useMemo(
    () => deals.filter((deal) => ["LookingForParticipants", "Discussion", "Confirmed"].includes(deal.status)),
    [deals],
  );

  useEffect(() => {
    if (actionableDeals.length === 0) {
      return;
    }

    const subscriptions = actionableDeals.map((deal) =>
      dispatch(dealsApi.endpoints.getDealJoins.initiate(deal.id)),
    );

    return () => {
      subscriptions.forEach((subscription) => subscription.unsubscribe());
    };
  }, [actionableDeals, dispatch]);

  const joinRequestsByDeal = useAppSelector((state) =>
    actionableDeals.map((deal) => ({
      deal,
      requests: dealsApi.endpoints.getDealJoins.select(deal.id)(state).data ?? [],
      isLoading:
        dealsApi.endpoints.getDealJoins.select(deal.id)(state).isLoading ||
        dealsApi.endpoints.getDealJoins.select(deal.id)(state).isUninitialized,
    })),
  );

  const dealsWithJoinRequests = useMemo(
    () =>
      joinRequestsByDeal
        .filter(({ requests }) => requests.length > 0)
        .map(({ deal, requests }) => ({
          deal,
          requests,
        })),
    [joinRequestsByDeal],
  );

  const joinRequestCount = useMemo(
    () => dealsWithJoinRequests.reduce((total, { requests }) => total + requests.length, 0),
    [dealsWithJoinRequests],
  );

  return {
    incomingDrafts,
    pendingReviews,
    dealsWithJoinRequests,
    draftCount: incomingDrafts.length,
    joinRequestCount,
    pendingReviewCount: pendingCount,
    totalActionCount: incomingDrafts.length + joinRequestCount + pendingCount,
    isLoading:
      isDraftsLoading ||
      isDealsLoading ||
      isPendingLoading ||
      joinRequestsByDeal.some(({ isLoading }) => isLoading),
    error: draftsError ?? dealsError ?? reviewsDealsError,
  };
}

export default useDealActionQueue;
