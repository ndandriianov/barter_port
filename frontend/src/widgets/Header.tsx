import {Link, useNavigate} from "react-router-dom";
import authApi from "@/features/auth/api/authApi.ts";
import {useAppDispatch} from "@/hooks/redux.ts";
import {logout} from "@/features/auth/model/authSlice.ts";

function Header() {
  const [apiLogout] = authApi.useLogoutMutation()
  const navigate = useNavigate()
  const dispatch = useAppDispatch();

  async function handleLogout() {
    await apiLogout();
    dispatch(logout());
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