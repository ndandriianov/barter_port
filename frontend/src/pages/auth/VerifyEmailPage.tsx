import { useSearchParams } from "react-router-dom";
import authApi from "@/features/auth/api/authApi";
import { useEffect, useState } from "react";
import { Alert, Box, CircularProgress, Typography } from "@mui/material";

type VerifyStatus = "pending" | "success" | "error";

function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") || "";

  const [verifyEmail, { isLoading }] = authApi.useVerifyEmailMutation();
  const [status, setStatus] = useState<VerifyStatus>("pending");

  useEffect(() => {
    async function verify() {
      if (token) {
        try {
          await verifyEmail({ token }).unwrap();
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
    });
  }, [token, verifyEmail]);

  return (
    <Box textAlign="center">
      <Typography variant="h5" fontWeight={700} mb={3}>
        Подтверждение email
      </Typography>

      {isLoading && (
        <Box display="flex" justifyContent="center" mt={2}>
          <CircularProgress />
        </Box>
      )}

      {!isLoading && status === "success" && (
        <Alert severity="success">Email успешно подтверждён!</Alert>
      )}

      {!isLoading && status === "error" && (
        <Alert severity="error">
          Недействительная или устаревшая ссылка подтверждения.
        </Alert>
      )}
    </Box>
  );
}

export default VerifyEmailPage;
