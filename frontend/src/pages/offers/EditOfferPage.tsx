import { Alert, Box, CircularProgress, Typography } from "@mui/material";
import { useParams } from "react-router-dom";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi";
import CreateOfferForm from "@/widgets/offers/CreateOfferForm";

function EditOfferPage() {
  const { offerId } = useParams<{ offerId: string }>();
  const { data: currentUser, isLoading: isUserLoading } = usersApi.useGetCurrentUserQuery();
  const { data: offer, isLoading: isOfferLoading, error } = offersApi.useGetOfferByIdQuery(offerId ?? "", {
    skip: !offerId,
  });

  if (!offerId) {
    return <Alert severity="warning">Объявление не найдено</Alert>;
  }

  if (isUserLoading || isOfferLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !offer) {
    return <Alert severity="warning">Объявление не найдено</Alert>;
  }

  if (!currentUser || currentUser.id !== offer.authorId) {
    return <Alert severity="warning">Редактировать объявление может только автор</Alert>;
  }

  return (
    <Box maxWidth={600} mx="auto">
      <Typography variant="h4" fontWeight={700} mb={3}>
        Редактирование объявления
      </Typography>
      <CreateOfferForm key={offer.id} mode="edit" offer={offer} />
    </Box>
  );
}

export default EditOfferPage;
