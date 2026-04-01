import {BrowserRouter, Routes, Route} from "react-router-dom";
import LoginPage from "@/pages/auth/LoginPage";
import ProfilePage from "@/pages/profile/ProfilePage";
import RegisterPage from "@/pages/auth/RegisterPage.tsx";
import VerifyEmailPage from "@/pages/auth/VerifyEmailPage.tsx";
import Header from "@/widgets/Header.tsx";
import CreateOfferPage from "@/pages/offers/CreateOfferPage.tsx";
import OffersListPage from "@/pages/offers/OffersListPage.tsx";
import OfferPage from "@/pages/offers/OfferPage.tsx";

function AppRouter() {
  return (
    <BrowserRouter>
      <Header />

      <Routes>
        <Route path="/login" element={<LoginPage/>}/>
        <Route path="/profile" element={<ProfilePage/>}/>
        <Route path="/register" element={<RegisterPage/>}/>
        <Route path="/verify-email" element={<VerifyEmailPage/>}/>
        <Route path="/offers" element={<OffersListPage/>}/>
        <Route path="/offers/create" element={<CreateOfferPage/>}/>
        <Route path="/offers/:offerId" element={<OfferPage/>}/>

        <Route path="*" element={<ProfilePage/>}/>
      </Routes>
    </BrowserRouter>
  )
}

export default AppRouter;
