import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

export function VerifyEmailPage() {
  const [params] = useSearchParams();
  const token = params.get("token");

  const [status, setStatus] = useState<"idle" | "loading" | "ok" | "error">(
    "idle"
  );
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function run() {
      if (!token) {
        setStatus("error");
        setError("Token is missing");
        return;
      }

      setStatus("loading");
      setError(null);

      const res = await fetch("http://localhost:8080/auth/verify-email", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token }),
      });

      const data = await res.json();

      if (!res.ok) {
        setStatus("error");
        setError(data?.error ?? "Unknown error");
        return;
      }

      setStatus("ok");
    }

    run();
  }, [token]);

  return (
    <div>
      <h1>Verify email</h1>

      {status === "loading" && <p>Verifying...</p>}
      {status === "ok" && <p>Success! Email verified.</p>}
      {status === "error" && <p>Error: {error}</p>}
    </div>
  );
}
