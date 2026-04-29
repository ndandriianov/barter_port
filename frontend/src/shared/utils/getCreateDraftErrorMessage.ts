import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type { SerializedError } from "@reduxjs/toolkit";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

type RequestError = FetchBaseQueryError | SerializedError | undefined;

export function getCreateDraftDealErrorMessage(error: RequestError): string | null {
  if (!error) {
    return null;
  }

  const backendMessage = getErrorMessage(error);
  const statusCode = getStatusCode(error);

  switch (statusCode) {
    case 400:
      return backendMessage ?? "Не удалось создать черновик: проверьте состав объявлений и количество.";
    case 401:
      return backendMessage ?? "Сессия истекла. Войдите снова и повторите попытку.";
    case 403:
      return backendMessage ?? "Нельзя создать черновик: один из авторов объявлений скрыл вас.";
    case 404:
      return backendMessage ?? "Одно или несколько объявлений не найдены.";
    case 500:
      return backendMessage ?? "Не удалось создать черновик из-за ошибки сервера.";
    default:
      return backendMessage ?? "Не удалось создать черновик сделки.";
  }
}

export function getCreateDraftFromOfferGroupErrorMessage(error: RequestError): string | null {
  if (!error) {
    return null;
  }

  const backendMessage = getErrorMessage(error);
  const statusCode = getStatusCode(error);

  switch (statusCode) {
    case 400:
      return backendMessage ?? "Не удалось создать черновик: проверьте выбор объявлений и условия отклика.";
    case 401:
      return backendMessage ?? "Сессия истекла. Войдите снова и повторите попытку.";
    case 403:
      return backendMessage ?? "Выбранное объявление для отклика вам не принадлежит.";
    case 404:
      return backendMessage ?? "Группа объявлений или выбранные объявления не найдены.";
    case 500:
      return backendMessage ?? "Не удалось создать черновик из-за ошибки сервера.";
    default:
      return backendMessage ?? "Не удалось создать черновик по группе объявлений.";
  }
}
