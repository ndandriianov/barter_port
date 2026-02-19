import {BrowserRouter, Routes, Route, Link} from "react-router-dom";
import {RegisterPage} from "./RegisterPage";
import {VerifyEmailPage} from "./VerifyEmailPage";
import {MePage} from "./MePage.tsx";
import {LoginPage} from "./LoginPage.tsx";

export default function App() {
  return (
    <BrowserRouter>
      <div>
        <nav>
          <Link to="/">Register</Link>
          <Link to={"/login"}>Login</Link>
        </nav>

        <MePage/>

        <Routes>
          <Route path="/" element={<RegisterPage/>}/>
          <Route path="/verify-email" element={<VerifyEmailPage/>}/>
          <Route path="/login" element={<LoginPage/>}/>
        </Routes>
      </div>
    </BrowserRouter>
  );
}
