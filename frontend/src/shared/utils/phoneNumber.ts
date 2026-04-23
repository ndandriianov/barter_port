const PHONE_NUMBER_REGEX = /^\+7 \(\d{3}\) \d{3}-\d{2}-\d{2}$/;

export function isValidPhoneNumber(value: string): boolean {
  return PHONE_NUMBER_REGEX.test(value.trim());
}

export function formatPhoneNumber(value?: string | null): string | undefined {
  if (!value) {
    return undefined;
  }

  const trimmed = value.trim();
  if (!trimmed) {
    return undefined;
  }

  if (PHONE_NUMBER_REGEX.test(trimmed)) {
    return trimmed;
  }

  let digits = trimmed.replace(/\D/g, "");
  if (!digits) {
    return undefined;
  }

  if (digits.length === 10 && digits.startsWith("9")) {
    digits = `7${digits}`;
  } else if (digits.length === 11 && digits.startsWith("8")) {
    digits = `7${digits.slice(1)}`;
  }

  if (digits.length !== 11 || !digits.startsWith("7")) {
    return trimmed;
  }

  return `+7 (${digits.slice(1, 4)}) ${digits.slice(4, 7)}-${digits.slice(7, 9)}-${digits.slice(9, 11)}`;
}

export function formatPhoneNumberInput(value: string): string {
  let digits = value.replace(/\D/g, "");
  if (!digits) {
    return "";
  }

  if (digits.startsWith("8")) {
    digits = `7${digits.slice(1)}`;
  } else if (!digits.startsWith("7")) {
    digits = `7${digits}`;
  }

  digits = digits.slice(0, 11);

  const code = digits.slice(1, 4);
  const part1 = digits.slice(4, 7);
  const part2 = digits.slice(7, 9);
  const part3 = digits.slice(9, 11);

  let result = "+7";
  if (code) {
    result += ` (${code}`;
  }
  if (code.length === 3) {
    result += ")";
  }
  if (part1) {
    result += ` ${part1}`;
  }
  if (part2) {
    result += `-${part2}`;
  }
  if (part3) {
    result += `-${part3}`;
  }

  return result;
}
