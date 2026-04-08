import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import AppLayout from "@/widgets/layout/AppLayout";
import AuthLayout from "@/widgets/layout/AuthLayout";
import LoginPage from "@/pages/auth/LoginPage";
import RegisterPage from "@/pages/auth/RegisterPage";
import VerifyEmailPage from "@/pages/auth/VerifyEmailPage";
import ProfilePage from "@/pages/profile/ProfilePage";
import CreateOfferPage from "@/pages/offers/CreateOfferPage";
import OffersListPage from "@/pages/offers/OffersListPage";
import OfferPage from "@/pages/offers/OfferPage";
import DealsListPage from "@/pages/deals/DealsListPage";
import DealPage from "@/pages/deals/DealPage";
import DraftsListPage from "@/pages/deals/DraftsListPage";
import DraftPage from "@/pages/deals/DraftPage";
import ChatsPage from "@/pages/chats/ChatsPage";

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
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/offers" element={<OffersListPage />} />
          <Route path="/offers/create" element={<CreateOfferPage />} />
          <Route path="/offers/:offerId" element={<OfferPage />} />
          <Route path="/deals" element={<DealsListPage />} />
          <Route path="/deals/:dealId" element={<DealPage />} />
          <Route path="/deals/drafts" element={<DraftsListPage />} />
          <Route path="/deals/drafts/:draftId" element={<DraftPage />} />
          <Route path="/chats" element={<ChatsPage />} />
          <Route path="*" element={<Navigate to="/profile" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default AppRouter;
