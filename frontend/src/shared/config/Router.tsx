import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import AppLayout from "@/widgets/layout/AppLayout";
import AuthLayout from "@/widgets/layout/AuthLayout";
import LoginPage from "@/pages/auth/LoginPage";
import RegisterPage from "@/pages/auth/RegisterPage";
import VerifyEmailPage from "@/pages/auth/VerifyEmailPage";
import AdminPage from "@/pages/admin/AdminPage";
import ProfilePage from "@/pages/profile/ProfilePage";
import CreateOfferPage from "@/pages/offers/CreateOfferPage";
import OffersListPage from "@/pages/offers/OffersListPage";
import OfferPage from "@/pages/offers/OfferPage";
import DealsListPage from "@/pages/deals/DealsListPage";
import DealPage from "@/pages/deals/DealPage";
import DraftsListPage from "@/pages/deals/DraftsListPage";
import DraftPage from "@/pages/deals/DraftPage";
import DealItemPage from "@/pages/deals/DealItemPage";
import ChatsPage from "@/pages/chats/ChatsPage";
import PendingReviewsPage from "@/pages/reviews/PendingReviewsPage";
import MyReviewsPage from "@/pages/reviews/MyReviewsPage";
import UserReviewsPage from "@/pages/reviews/UserReviewsPage";
import OfferReviewsPage from "@/pages/reviews/OfferReviewsPage";

function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AuthLayout />}>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/verify-email" element={<VerifyEmailPage />} />
        </Route>

        <Route element={<AppLayout />}>
          <Route path="/admin" element={<AdminPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/offers" element={<OffersListPage />} />
          <Route path="/offers/create" element={<CreateOfferPage />} />
          <Route path="/offers/:offerId" element={<OfferPage />} />
          <Route path="/offers/:offerId/reviews" element={<OfferReviewsPage />} />
          <Route path="/deals" element={<DealsListPage />} />
          <Route path="/deals/:dealId" element={<DealPage />} />
          <Route path="/deals/:dealId/items/:itemId" element={<DealItemPage />} />
          <Route path="/deals/drafts" element={<DraftsListPage />} />
          <Route path="/deals/drafts/:draftId" element={<DraftPage />} />
          <Route path="/chats" element={<ChatsPage />} />
          <Route path="/reviews/pending" element={<PendingReviewsPage />} />
          <Route path="/reviews/mine" element={<MyReviewsPage />} />
          <Route path="/users/:userId/reviews" element={<UserReviewsPage />} />
          <Route path="*" element={<Navigate to="/profile" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default AppRouter;
