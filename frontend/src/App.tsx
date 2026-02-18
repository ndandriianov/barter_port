import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import { RegisterPage } from "./RegisterPage";
import { VerifyEmailPage } from "./VerifyEmailPage";

export default function App() {
  return (
    <BrowserRouter>
      <div>
        <nav>
          <Link to="/">Register</Link>
        </nav>

        <Routes>
          <Route path="/" element={<RegisterPage />} />
          <Route path="/verify-email" element={<VerifyEmailPage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}
