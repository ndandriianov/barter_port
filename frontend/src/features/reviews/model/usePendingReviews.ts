import { useEffect, useMemo } from "react";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";

interface UsePendingReviewsOptions {
  selectedDealId?: string | null;
}

function usePendingReviews({ selectedDealId }: UsePendingReviewsOptions = {}) {
  const dispatch = useAppDispatch();
  const { data: deals, isLoading: isDealsLoading, error: dealsError } = dealsApi.useGetDealsQuery({ my: true });

  const completedDeals = useMemo(
    () => (deals ?? []).filter((deal) => deal.status === "Completed"),
    [deals],
  );

  useEffect(() => {
    if (completedDeals.length === 0) {
      return;
    }

    const subscriptions = completedDeals.map((deal) =>
      dispatch(reviewsApi.endpoints.getDealPendingReviews.initiate(deal.id)),
    );

    return () => {
      subscriptions.forEach((subscription) => subscription.unsubscribe());
    };
  }, [completedDeals, dispatch]);

  const pendingByDeal = useAppSelector((state) =>
    completedDeals.map((deal) => ({
      deal,
      query: reviewsApi.endpoints.getDealPendingReviews.select(deal.id)(state),
    })),
  );

  const pendingReviews = useMemo(
    () =>
      pendingByDeal.flatMap(({ deal, query }) =>
        (query.data ?? [])
          .filter((review) => review.canCreate)
          .map((review) => ({ deal, review })),
      ),
    [pendingByDeal],
  );

  const filteredReviews = useMemo(
    () =>
      selectedDealId
        ? pendingReviews.filter(({ deal }) => deal.id === selectedDealId)
        : pendingReviews,
    [pendingReviews, selectedDealId],
  );

  const isPendingLoading =
    completedDeals.length > 0 &&
    pendingByDeal.some(({ query }) => query.isLoading || query.isUninitialized);
  const hasPendingErrors = pendingByDeal.some(({ query }) => query.isError);

  return {
    completedDeals,
    pendingReviews,
    filteredReviews,
    pendingCount: pendingReviews.length,
    isDealsLoading,
    dealsError,
    isPendingLoading,
    hasPendingErrors,
  };
}

export default usePendingReviews;
