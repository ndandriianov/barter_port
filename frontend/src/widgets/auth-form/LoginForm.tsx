import {useState} from "react";
import authApi from "@/features/auth/api/authApi";
import {useNavigate} from "react-router-dom";

function LoginForm() {
  const [login, {isLoading, error}] = authApi.useLoginMutation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    await login({email, password});
    navigate("/");
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
        Login
      </button>
      {error && <div>Login error</div>}
    </form>
  );
}

export default LoginForm;