const MAX_AVATAR_SIZE = 512;
const AVATAR_MIME_TYPE = "image/webp";
const AVATAR_QUALITY = 0.86;

function loadImage(file: File): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    const objectUrl = URL.createObjectURL(file);

    image.onload = () => {
      URL.revokeObjectURL(objectUrl);
      resolve(image);
    };

    image.onerror = () => {
      URL.revokeObjectURL(objectUrl);
      reject(new Error("Не удалось прочитать изображение."));
    };

    image.src = objectUrl;
  });
}

export async function imageToAvatarDataUrl(file: File): Promise<string> {
  const image = await loadImage(file);
  const longestSide = Math.max(image.width, image.height, 1);
  const scale = Math.min(1, MAX_AVATAR_SIZE / longestSide);
  const width = Math.max(1, Math.round(image.width * scale));
  const height = Math.max(1, Math.round(image.height * scale));

  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = height;

  const context = canvas.getContext("2d");
  if (!context) {
    throw new Error("Не удалось подготовить изображение.");
  }

  context.drawImage(image, 0, 0, width, height);

  return canvas.toDataURL(AVATAR_MIME_TYPE, AVATAR_QUALITY);
}
