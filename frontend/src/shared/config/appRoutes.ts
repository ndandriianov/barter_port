export type AppSection = "market" | "deals" | "messages" | "profile" | "admin";

export const appRoutes = {
  auth: {
    login: "/login",
    register: "/register",
    verifyEmail: "/verify-email",
    resetPassword: "/reset-password",
  },
  market: {
    home: "/app/market",
    catalog: "/app/market/catalog",
    catalogSubscriptions: "/app/market/catalog/subscriptions",
    catalogFavorites: "/app/market/catalog/favorites",
    catalogMine: "/app/market/catalog/mine",
    createOffer: "/app/market/offers/create",
    offer: (offerId: string) => `/app/market/offers/${offerId}`,
    editOffer: (offerId: string) => `/app/market/offers/${offerId}/edit`,
    offerReviews: (offerId: string) => `/app/market/offers/${offerId}/reviews`,
    exchangeGroups: "/app/market/exchange-groups",
    exchangeGroupsMine: "/app/market/exchange-groups/mine",
    createExchangeGroup: "/app/market/exchange-groups/create",
    exchangeGroup: (offerGroupId: string) => `/app/market/exchange-groups/${offerGroupId}`,
    myPublications: "/app/market/my-publications",
    myPublicationOffers: "/app/market/my-publications/offers",
    myPublicationGroups: "/app/market/my-publications/groups",
    myPublicationModeration: "/app/market/my-publications/moderation",
  },
  deals: {
    home: "/app/deals",
    tasks: "/app/deals/tasks",
    drafts: "/app/deals/drafts",
    draftsIncoming: "/app/deals/drafts/incoming",
    draftsMine: "/app/deals/drafts/mine",
    active: "/app/deals/active",
    history: "/app/deals/history",
    detail: (dealId: string) => `/app/deals/${dealId}`,
    item: (dealId: string, itemId: string) => `/app/deals/${dealId}/items/${itemId}`,
    draftDetail: (draftId: string) => `/app/deals/drafts/${draftId}`,
    reviews: "/app/deals/reviews",
  },
  messages: {
    home: "/app/messages",
    direct: "/app/messages/direct",
    deal: "/app/messages/deals",
  },
  profile: {
    home: "/app/profile",
    account: "/app/profile/account",
    accountPassword: "/app/profile/account/password",
    reputation: "/app/profile/reputation",
    reputationHistory: "/app/profile/reputation/history",
    networkSubscriptions: "/app/profile/network/subscriptions",
    networkSubscribers: "/app/profile/network/subscribers",
    reviewsMine: "/app/profile/reviews",
    reviewsAboutMe: "/app/profile/reviews/about-me",
    statistics: "/app/profile/statistics",
  },
  admin: {
    home: "/app/admin",
    offerReports: "/app/admin/offer-reports",
    offerReport: (reportId: string) => `/app/admin/offer-reports/${reportId}`,
    failures: "/app/admin/failures",
    system: "/app/admin/system",
  },
} as const;

export function getSectionFromPathname(pathname: string): AppSection {
  if (pathname.startsWith("/app/deals")) {
    return "deals";
  }

  if (pathname.startsWith("/app/messages")) {
    return "messages";
  }

  if (pathname.startsWith("/app/profile")) {
    return "profile";
  }

  if (pathname.startsWith("/app/admin")) {
    return "admin";
  }

  return "market";
}
