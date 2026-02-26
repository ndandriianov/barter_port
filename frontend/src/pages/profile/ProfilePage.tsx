import authApi from "@/features/auth/api/authApi";
import {useAppDispatch} from "@/hooks/redux.ts";
import {useNavigate} from "react-router-dom";
import {performLogout} from "@/features/auth/model/logoutThunk.ts";

function ProfilePage() {
  const {data, isLoading, refetch} = authApi.useMeQuery();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  if (isLoading) return <div>Loading...</div>;
  if (!data) return <div>Not authorized</div>;

  async function handleLogout() {
    await dispatch(performLogout());
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