import authApi from "@/features/auth/api/authApi";
import {useAppDispatch} from "@/hooks/redux.ts";
import {logout} from "@/features/auth/model/authSlice.ts";
import {useNavigate} from "react-router-dom";

function ProfilePage() {
  const { data, isLoading, refetch } = authApi.useMeQuery();
  const [apiLogout] = authApi.useLogoutMutation();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  if (isLoading) return <div>Loading...</div>;
  if (!data) return <div>Not authorized</div>;

  async function handleLogout() {
    await apiLogout();
    dispatch(logout());
    navigate("/login");
  }

  return (
    <div>
      <div>UserID: {data.userId}</div>
      <button onClick={handleLogout}>Logout</button>
      <button onClick={() => refetch()}>Refetch</button>
    </div>
  );
}

export default ProfilePage;