export interface LatLon {
  lat: number;
  lon: number;
}

declare global {
  interface Window {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ymaps?: any;
    __ymapsPromise?: Promise<void>;
  }
}

export const DEFAULT_CENTER: [number, number] = [55.751244, 37.618423]; // Moscow

export function loadYmaps(apiKey: string): Promise<void> {
  if (window.ymaps?.ready) return Promise.resolve();
  if (window.__ymapsPromise) return window.__ymapsPromise;

  window.__ymapsPromise = new Promise<void>((resolve, reject) => {
    const script = document.createElement("script");
    script.src = `https://api-maps.yandex.ru/2.1/?apikey=${apiKey}&lang=ru_RU`;
    script.async = true;
    script.onload = () => {
      window.ymaps.ready(() => resolve());
    };
    script.onerror = () => reject(new Error("Не удалось загрузить Яндекс Карты"));
    document.head.appendChild(script);
  });

  return window.__ymapsPromise;
}
