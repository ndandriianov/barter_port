import { useEffect, useState } from "react";

type MeResponse = {
  userId: string;
};

export function MePage() {
  const [data, setData] = useState<MeResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchMe() {
      try {
        const token = localStorage.getItem("accessToken");

        if (!token) {
          setError("no token found");
          setLoading(false);
          return;
        }

        const res = await fetch("http://localhost:8080/auth/me", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        const body = await res.json();

        if (!res.ok) {
          setError(body.error || "request failed");
          setLoading(false);
          return;
        }

        setData(body);
      } catch (e) {
        setError("network error");
      } finally {
        setLoading(false);
      }
    }

    fetchMe();
  }, []);

  return (
    <div>
      <h2>Me</h2>

      {loading && <p>loading...</p>}

      {error && <p>error: {error}</p>}

      {data && (
        <div>
          <p>UserID: {data.userId}</p>
        </div>
      )}
    </div>
  );
}
