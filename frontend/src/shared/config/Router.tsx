import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
  useParams,
  useSearchParams,
} from "react-router-dom";
import AppLayout from "@/widgets/layout/AppLayout";
import AuthLayout from "@/widgets/layout/AuthLayout";
import LoginPage from "@/pages/auth/LoginPage";
import RegisterPage from "@/pages/auth/RegisterPage";
import VerifyEmailPage from "@/pages/auth/VerifyEmailPage";
import ResetPasswordPage from "@/pages/auth/ResetPasswordPage.tsx";
import AdminPage from "@/pages/admin/AdminPage";
import AdminOfferReportsPage from "@/pages/admin/AdminOfferReportsPage.tsx";
import AdminOfferReportDetailsPage from "@/pages/admin/AdminOfferReportDetailsPage.tsx";
import ModerationHomePage from "@/pages/admin/ModerationHomePage.tsx";
import FailureModerationPage from "@/pages/admin/FailureModerationPage.tsx";
import ProfilePage from "@/pages/profile/ProfilePage";
import ProfileHomePage from "@/pages/profile/ProfileHomePage.tsx";
import ProfilePasswordPage from "@/pages/profile/ProfilePasswordPage.tsx";
import ProfileReputationPage from "@/pages/profile/ProfileReputationPage.tsx";
import ProfileNetworkPage from "@/pages/profile/ProfileNetworkPage.tsx";
import ProfileReviewsPage from "@/pages/profile/ProfileReviewsPage.tsx";
import CreateOfferPage from "@/pages/offers/CreateOfferPage";
import EditOfferPage from "@/pages/offers/EditOfferPage";
import OffersListPage from "@/pages/offers/OffersListPage";
import OfferPage from "@/pages/offers/OfferPage";
import MyOfferReportsPage from "@/pages/offers/MyOfferReportsPage.tsx";
import OfferGroupsListPage from "@/pages/offer-groups/OfferGroupsListPage.tsx";
import CreateOfferGroupPage from "@/pages/offer-groups/CreateOfferGroupPage.tsx";
import OfferGroupPage from "@/pages/offer-groups/OfferGroupPage.tsx";
import DealPage from "@/pages/deals/DealPage";
import DraftsListPage from "@/pages/deals/DraftsListPage";
import DraftPage from "@/pages/deals/DraftPage";
import DealItemPage from "@/pages/deals/DealItemPage";
import DealsHomePage from "@/pages/deals/DealsHomePage.tsx";
import DealTasksPage from "@/pages/deals/DealTasksPage.tsx";
import DealsStatusBoardPage from "@/pages/deals/DealsStatusBoardPage.tsx";
import ChatsPage from "@/pages/chats/ChatsPage";
import ReviewsPage from "@/pages/reviews/ReviewsPage";
import UserReviewsPage from "@/pages/reviews/UserReviewsPage";
import OfferReviewsPage from "@/pages/reviews/OfferReviewsPage";
import UserPage from "@/pages/users/UserPage.tsx";
import StatisticsPage from "@/pages/statistics/StatisticsPage.tsx";
import MarketHomePage from "@/pages/market/MarketHomePage.tsx";
import MarketCatalogPage from "@/pages/market/MarketCatalogPage.tsx";
import MarketOfferGroupsPage from "@/pages/market/MarketOfferGroupsPage.tsx";
import MyPublicationsPage from "@/pages/market/MyPublicationsPage.tsx";
import { appRoutes } from "@/shared/config/appRoutes.ts";

function LegacyOffersRedirect() {
  const [searchParams] = useSearchParams();
  const tab = searchParams.get("tab");

  if (tab === "favorites") {
    return <Navigate to={appRoutes.market.catalogFavorites} replace />;
  }

  if (tab === "subscriptions") {
    return <Navigate to={appRoutes.market.catalogSubscriptions} replace />;
  }

  if (tab === "mine") {
    return <Navigate to={appRoutes.market.myPublicationOffers} replace />;
  }

  return <Navigate to={appRoutes.market.catalog} replace />;
}

function LegacyOfferGroupsRedirect() {
  const [searchParams] = useSearchParams();
  const tab = searchParams.get("tab");

  return <Navigate to={tab === "mine" ? appRoutes.market.exchangeGroupsMine : appRoutes.market.exchangeGroups} replace />;
}

function LegacyDraftsRedirect() {
  const [searchParams] = useSearchParams();
  const tab = searchParams.get("tab");

  if (tab === "others") {
    return <Navigate to={appRoutes.deals.draftsIncoming} replace />;
  }

  if (tab === "mine") {
    return <Navigate to={appRoutes.deals.draftsMine} replace />;
  }

  return <Navigate to={appRoutes.deals.drafts} replace />;
}

function LegacyReviewsRedirect() {
  const [searchParams] = useSearchParams();
  const tab = searchParams.get("tab");
  const dealId = searchParams.get("dealId");

  if (tab === "mine") {
    return <Navigate to={appRoutes.profile.reviewsMine} replace />;
  }

  if (tab === "about-me") {
    return <Navigate to={appRoutes.profile.reviewsAboutMe} replace />;
  }

  const destination = dealId
    ? `${appRoutes.deals.reviews}?dealId=${encodeURIComponent(dealId)}`
    : appRoutes.deals.reviews;

  return <Navigate to={destination} replace />;
}

function RedirectLegacyOffer() {
  const { offerId } = useParams<{ offerId: string }>();
  return offerId ? <Navigate to={appRoutes.market.offer(offerId)} replace /> : <Navigate to={appRoutes.market.catalog} replace />;
}

function RedirectLegacyOfferEdit() {
  const { offerId } = useParams<{ offerId: string }>();
  return offerId ? <Navigate to={appRoutes.market.editOffer(offerId)} replace /> : <Navigate to={appRoutes.market.catalog} replace />;
}

function RedirectLegacyOfferReviews() {
  const { offerId } = useParams<{ offerId: string }>();
  return offerId ? <Navigate to={appRoutes.market.offerReviews(offerId)} replace /> : <Navigate to={appRoutes.market.catalog} replace />;
}

function RedirectLegacyOfferGroup() {
  const { offerGroupId } = useParams<{ offerGroupId: string }>();
  return offerGroupId ? <Navigate to={appRoutes.market.exchangeGroup(offerGroupId)} replace /> : <Navigate to={appRoutes.market.exchangeGroups} replace />;
}

function RedirectLegacyDeal() {
  const { dealId } = useParams<{ dealId: string }>();
  return dealId ? <Navigate to={appRoutes.deals.detail(dealId)} replace /> : <Navigate to={appRoutes.deals.home} replace />;
}

function RedirectLegacyDealItem() {
  const { dealId, itemId } = useParams<{ dealId: string; itemId: string }>();
  return dealId && itemId
    ? <Navigate to={appRoutes.deals.item(dealId, itemId)} replace />
    : <Navigate to={appRoutes.deals.home} replace />;
}

function RedirectLegacyDraft() {
  const { draftId } = useParams<{ draftId: string }>();
  return draftId ? <Navigate to={appRoutes.deals.draftDetail(draftId)} replace /> : <Navigate to={appRoutes.deals.drafts} replace />;
}

function RedirectLegacyAdminReport() {
  const { reportId } = useParams<{ reportId: string }>();
  return reportId ? <Navigate to={appRoutes.admin.offerReport(reportId)} replace /> : <Navigate to={appRoutes.admin.offerReports} replace />;
}

function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AuthLayout />}>
          <Route path={appRoutes.auth.login} element={<LoginPage />} />
          <Route path={appRoutes.auth.register} element={<RegisterPage />} />
          <Route path={appRoutes.auth.verifyEmail} element={<VerifyEmailPage />} />
          <Route path={appRoutes.auth.resetPassword} element={<ResetPasswordPage />} />
        </Route>

        <Route element={<AppLayout />}>
          <Route path="/" element={<Navigate to={appRoutes.market.home} replace />} />
          <Route path="/app" element={<Navigate to={appRoutes.market.home} replace />} />

          <Route path={appRoutes.market.home} element={<MarketHomePage />} />
          <Route path={appRoutes.market.catalog} element={<MarketCatalogPage mode="others" />} />
          <Route path={appRoutes.market.catalogSubscriptions} element={<MarketCatalogPage mode="subscriptions" />} />
          <Route path={appRoutes.market.catalogFavorites} element={<MarketCatalogPage mode="favorites" />} />
          <Route
            path={appRoutes.market.myPublicationOffers}
            element={
              <OffersListPage
                forcedTab="mine"
                hideTabs
                title="Мои объявления"
                description=""
              />
            }
          />
          <Route path={appRoutes.market.createOffer} element={<CreateOfferPage />} />
          <Route path="/app/market/offers/:offerId/edit" element={<EditOfferPage />} />
          <Route path="/app/market/offers/:offerId/reviews" element={<OfferReviewsPage />} />
          <Route path="/app/market/offers/:offerId" element={<OfferPage />} />
          <Route path={appRoutes.market.exchangeGroups} element={<MarketOfferGroupsPage mode="others" />} />
          <Route path={appRoutes.market.exchangeGroupsMine} element={<MarketOfferGroupsPage mode="mine" />} />
          <Route
            path={appRoutes.market.myPublicationGroups}
            element={
              <OfferGroupsListPage
                forcedTab="mine"
                hideTabs
                title="Мои сценарии обмена"
                description="Ваши composite offer-group публикации в зоне управления собственными материалами."
              />
            }
          />
          <Route path={appRoutes.market.createExchangeGroup} element={<CreateOfferGroupPage />} />
          <Route path="/app/market/exchange-groups/:offerGroupId" element={<OfferGroupPage />} />
          <Route path={appRoutes.market.myPublications} element={<MyPublicationsPage />} />
          <Route path={appRoutes.market.myPublicationModeration} element={<MyOfferReportsPage />} />

          <Route path={appRoutes.deals.home} element={<DealsHomePage />} />
          <Route path={appRoutes.deals.tasks} element={<DealTasksPage />} />
          <Route path={appRoutes.deals.drafts} element={<DraftsListPage />} />
          <Route
            path={appRoutes.deals.draftsIncoming}
            element={
              <DraftsListPage
                forcedTab="others"
                hideTabs
                title="Входящие черновики"
                description="Предложения других пользователей, которые требуют подтверждения или отклонения."
              />
            }
          />
          <Route
            path={appRoutes.deals.draftsMine}
            element={
              <DraftsListPage
                forcedTab="mine"
                hideTabs
                title="Исходящие черновики"
                description="Ваши собственные draft-сценарии до момента перехода в сделку."
              />
            }
          />
          <Route path="/app/deals/drafts/:draftId" element={<DraftPage />} />
          <Route path={appRoutes.deals.active} element={<DealsStatusBoardPage mode="active" />} />
          <Route path={appRoutes.deals.history} element={<DealsStatusBoardPage mode="history" />} />
          <Route
            path={appRoutes.deals.reviews}
            element={
              <ReviewsPage
                forcedTab="available"
                hideTabs
                hideBackButton
                title="Отзывы после завершения"
                description="Доступные отзывы встроены в lifecycle сделки и больше не живут отдельным top-level разделом."
              />
            }
          />
          <Route path="/app/deals/:dealId/items/:itemId" element={<DealItemPage />} />
          <Route path="/app/deals/:dealId" element={<DealPage />} />

          <Route path={appRoutes.messages.home} element={<ChatsPage />} />
          <Route path={appRoutes.messages.direct} element={<ChatsPage defaultMode="direct" />} />
          <Route path={appRoutes.messages.deal} element={<ChatsPage defaultMode="deal" />} />

          <Route path={appRoutes.profile.home} element={<ProfileHomePage />} />
          <Route path={appRoutes.profile.account} element={<ProfilePage />} />
          <Route path={appRoutes.profile.accountPassword} element={<ProfilePasswordPage />} />
          <Route path={appRoutes.profile.reputation} element={<ProfileReputationPage />} />
          <Route path={appRoutes.profile.reputationHistory} element={<ProfileReputationPage historyMode />} />
          <Route
            path={appRoutes.profile.networkSubscriptions}
            element={<ProfileNetworkPage mode="subscriptions" />}
          />
          <Route
            path={appRoutes.profile.networkSubscribers}
            element={<ProfileNetworkPage mode="subscribers" />}
          />
          <Route
            path={appRoutes.profile.reviewsMine}
            element={
              <ProfileReviewsPage mode="mine" />
            }
          />
          <Route
            path={appRoutes.profile.reviewsAboutMe}
            element={
              <ProfileReviewsPage mode="about-me" />
            }
          />
          <Route path={appRoutes.profile.statistics} element={<StatisticsPage />} />

          <Route path={appRoutes.admin.home} element={<ModerationHomePage />} />
          <Route path={appRoutes.admin.offerReports} element={<AdminOfferReportsPage />} />
          <Route path="/app/admin/offer-reports/:reportId" element={<AdminOfferReportDetailsPage />} />
          <Route path={appRoutes.admin.failures} element={<FailureModerationPage />} />
          <Route path={appRoutes.admin.system} element={<AdminPage />} />

          <Route path="/users/:userId" element={<UserPage />} />
          <Route path="/users/:userId/reviews" element={<UserReviewsPage />} />

          <Route path="/admin" element={<Navigate to={appRoutes.admin.home} replace />} />
          <Route path="/admin/offer-reports" element={<Navigate to={appRoutes.admin.offerReports} replace />} />
          <Route path="/admin/offer-reports/:reportId" element={<RedirectLegacyAdminReport />} />
          <Route path="/profile" element={<Navigate to={appRoutes.profile.account} replace />} />
          <Route path="/offer-reports/mine" element={<Navigate to={appRoutes.market.myPublicationModeration} replace />} />
          <Route path="/offers" element={<LegacyOffersRedirect />} />
          <Route path="/offers/create" element={<Navigate to={appRoutes.market.createOffer} replace />} />
          <Route path="/offers/:offerId/edit" element={<RedirectLegacyOfferEdit />} />
          <Route path="/offers/:offerId/reviews" element={<RedirectLegacyOfferReviews />} />
          <Route path="/offers/:offerId" element={<RedirectLegacyOffer />} />
          <Route path="/offer-groups" element={<LegacyOfferGroupsRedirect />} />
          <Route path="/offer-groups/create" element={<Navigate to={appRoutes.market.createExchangeGroup} replace />} />
          <Route path="/offer-groups/:offerGroupId" element={<RedirectLegacyOfferGroup />} />
          <Route path="/deals/drafts/:draftId" element={<RedirectLegacyDraft />} />
          <Route path="/deals/drafts" element={<LegacyDraftsRedirect />} />
          <Route path="/deals/:dealId/items/:itemId" element={<RedirectLegacyDealItem />} />
          <Route path="/deals/:dealId" element={<RedirectLegacyDeal />} />
          <Route path="/deals" element={<Navigate to={appRoutes.deals.home} replace />} />
          <Route path="/chats" element={<Navigate to={appRoutes.messages.home} replace />} />
          <Route path="/reviews" element={<LegacyReviewsRedirect />} />
          <Route path="/reviews/pending" element={<Navigate to={appRoutes.deals.reviews} replace />} />
          <Route path="/reviews/mine" element={<Navigate to={appRoutes.profile.reviewsMine} replace />} />
          <Route path="/statistics" element={<Navigate to={appRoutes.profile.statistics} replace />} />
          <Route path="*" element={<Navigate to={appRoutes.market.home} replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default AppRouter;
