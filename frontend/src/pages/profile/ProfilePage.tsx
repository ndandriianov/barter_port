import { useMeQuery, useLogoutMutation } from "@/features/auth/api/authApi";

function ProfilePage() {
  const { data, isLoading } = useMeQuery();
  const [logout] = useLogoutMutation();

  if (isLoading) return <div>Loading...</div>;
  if (!data) return <div>Not authorized</div>;

  return (
    <div>
      <div>UserID: {data.userId}</div>
      <button onClick={() => logout()}>Logout</button>
    </div>
  );
}

export default ProfilePage;