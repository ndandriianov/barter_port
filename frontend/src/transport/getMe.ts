export async function getMe() {
  const token = localStorage.getItem("accessToken");

  const res = await fetch("http://localhost:8080/auth/me", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return res.json();
}
