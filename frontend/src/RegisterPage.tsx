import { useState } from "react";

type RegisterResponse = {
  userId: string;
  email: string;
  verifyUrl: string;
};

export function RegisterPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const [result, setResult] = useState<RegisterResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setResult(null);

    const res = await fetch("http://localhost:8080/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    const data = await res.json();

    if (!res.ok) {
      setError(data?.error ?? "Unknown error");
      return;
    }

    setResult(data);
  }

  return (
    <div>
      <h1>Register</h1>

      <form onSubmit={onSubmit}>
        <div>
          <label>Email</label>
          <br />
          <input value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>

        <div>
          <label>Password</label>
          <br />
          <input
            value={password}
            type="password"
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>

        <button type="submit">Create account</button>
      </form>

      {error && (
        <div>
          <p>Error: {error}</p>
        </div>
      )}

      {result && (
        <div>
          <p>Registered: {result.email}</p>
          <p>
            Verification link (temporary):{" "}
            <a href={result.verifyUrl}>{result.verifyUrl}</a>
          </p>
        </div>
      )}
    </div>
  );
}
