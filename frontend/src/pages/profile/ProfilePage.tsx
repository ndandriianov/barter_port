import authApi from "@/features/auth/api/authApi";

function ProfilePage() {
  const { data, isLoading, refetch } = authApi.useMeQuery();
  const [logout] = authApi.useLogoutMutation();

  if (isLoading) return <div>Loading...</div>;
  if (!data) return <div>Not authorized</div>;

  return (
    <div>
      <div>UserID: {data.userId}</div>
      <button onClick={() => logout()}>Logout</button>
      <button onClick={() => refetch()}>Refetch</button>
    </div>
  );
}

export default ProfilePage;