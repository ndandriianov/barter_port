import {Link, useNavigate} from "react-router-dom";
import {useAppDispatch} from "@/hooks/redux.ts";
import {performLogout} from "@/features/auth/model/logoutThunk.ts";

function Header() {
  const navigate = useNavigate()
  const dispatch = useAppDispatch();

  async function handleLogout() {
    await dispatch(performLogout());
    navigate("/login");
  }

  return (
    <div>
      <Link to={"/login"}>Login</Link>
      <Link to={"/profile"}>Profile</Link>
      <Link to={"/register"}>Register</Link>
      <Link to={"/items"}>Items</Link>
      <Link to={"/items/create"}>Create Item</Link>
      <button onClick={handleLogout}>Logout</button>
    </div>
  );
}

export default Header;
