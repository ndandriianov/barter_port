import {BrowserRouter, Routes, Route, Link} from "react-router-dom";
import LoginPage from "@/pages/auth/LoginPage";
import ProfilePage from "@/pages/profile/ProfilePage";
import RegisterPage from "@/pages/auth/RegisterPage.tsx";
import VerifyEmailPage from "@/pages/auth/VerifyEmailPage.tsx";

function AppRouter() {
  return (
    <BrowserRouter>
      <Link to={"/profile"}>Profile</Link>

      <Routes>
        <Route path="/login" element={<LoginPage/>}/>
        <Route path="/profile" element={<ProfilePage/>}/>
        <Route path="/register" element={<RegisterPage/>}/>
        <Route path="/verify-email" element={<VerifyEmailPage/>}/>
      </Routes>
    </BrowserRouter>
  )
}

export default AppRouter;