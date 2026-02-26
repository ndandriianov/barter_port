import {useSearchParams} from "react-router-dom";
import authApi from "@/features/auth/api/authApi.ts";
import {useEffect, useState} from "react";

type verifyStatus = "pending" | "success" | "error"

function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") || "";

  const [verifyEmail, {isLoading}] = authApi.useVerifyEmailMutation();
  const [status, setStatus] = useState<verifyStatus>("pending")

  useEffect(() => {
    async function verify() {
      if (token) {
        try {
          await verifyEmail({token}).unwrap();
          setStatus("success");
        } catch {
          setStatus("error");
        }
      } else {
        setStatus("error");
      }
    }

    verify().catch((error) => {
      console.error("Error verifying email:", error);
    })
  }, [token, verifyEmail]);

  if (isLoading) {
    return <div>Verifying your email...</div>;
  }

  if (status === "success") {
    return <div>Your email has been successfully verified!</div>;
  }

  if (status === "error") {
    return <div>Invalid or expired verification link. Please request a new one.</div>;
  }

  return (
    <div></div>
  );
}

export default VerifyEmailPage;