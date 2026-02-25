import {useState} from "react";
import {useRegisterMutation} from "@/features/auth/api/authApi";

function RegisterForm() {
  const [register, { isLoading, error }] = useRegisterMutation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    await register({ email, password });
  };

  return (
    <form onSubmit={submit}>
      <input
        placeholder="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      <input
        placeholder="password"
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      <button type="submit" disabled={isLoading}>
        Register
      </button>
      {error && <div>Login error</div>}
    </form>
  );
}

export default RegisterForm;