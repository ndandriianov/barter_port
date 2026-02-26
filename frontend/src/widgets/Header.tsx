import {Link, useNavigate} from "react-router-dom";
import authApi from "@/features/auth/api/authApi.ts";

function Header() {
  const [logout] = authApi.useLogoutMutation()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout();
    navigate("/login");
  }

  return (
    <div>
      <Link to={"/login"}>Login</Link>
      <Link to={"/profile"}>Profile</Link>
      <Link to={"/register"}>Register</Link>
      <button onClick={handleLogout}>Logout</button>
    </div>
  );
}

export default Header;