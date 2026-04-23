import { useEffect, useRef, useState } from "react";

export interface LatLon {
  lat: number;
  lon: number;
}

interface Props {
  value: LatLon | null;
  onChange: (value: LatLon | null) => void;
  height?: string;
  /** Optional API key. Falls back to VITE_YANDEX_MAPS_API_KEY env variable. */
  apiKey?: string;
}

declare global {
  interface Window {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ymaps3?: any;
    __ymaps3Promise?: Promise<void>;
  }
}

const DEFAULT_CENTER: [number, number] = [55.751244, 37.618423]; // Moscow

function loadYmaps(apiKey: string): Promise<void> {
  if (window.ymaps3) return Promise.resolve();
  if (window.__ymaps3Promise) return window.__ymaps3Promise;

  window.__ymaps3Promise = new Promise<void>((resolve, reject) => {
    const script = document.createElement("script");
    script.src = `https://api-maps.yandex.ru/3.0/?apikey=${apiKey}&lang=ru_RU`;
    script.async = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("Failed to load Yandex Maps API"));
    document.head.appendChild(script);
  });

  return window.__ymaps3Promise;
}

export default function YandexMapPicker({ value, onChange, height = "300px", apiKey }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<unknown>(null);
  const markerRef = useRef<unknown>(null);
  const [error, setError] = useState<string | null>(null);
  const [loaded, setLoaded] = useState(false);

  const resolvedApiKey = apiKey ?? import.meta.env.VITE_YANDEX_MAPS_API_KEY ?? "";

  useEffect(() => {
    if (!resolvedApiKey) {
      setError("Яндекс Карты: API-ключ не задан (VITE_YANDEX_MAPS_API_KEY)");
      return;
    }

    let cancelled = false;

    loadYmaps(resolvedApiKey)
      .then(async () => {
        if (cancelled || !containerRef.current) return;

        const ymaps3 = window.ymaps3!;
        await ymaps3.ready;

        if (cancelled || !containerRef.current) return;

        const { YMap, YMapDefaultSchemeLayer, YMapDefaultFeaturesLayer, YMapMarker } = ymaps3;

        const center: [number, number] = value ? [value.lat, value.lon] : DEFAULT_CENTER;

        const map = new YMap(containerRef.current, {
          location: { center, zoom: value ? 14 : 10 },
        });

        map.addChild(new YMapDefaultSchemeLayer({}));
        map.addChild(new YMapDefaultFeaturesLayer({}));

        mapRef.current = map;

        // Place initial marker
        if (value) {
          const el = createMarkerElement();
          const marker = new YMapMarker({ coordinates: [value.lat, value.lon] }, el);
          map.addChild(marker);
          markerRef.current = marker;
        }

        map.addChild({
          // Pseudo-listener layer for click events
          onUpdate() {},
          onAttach() {},
          onDetach() {},
        });

        containerRef.current.addEventListener("click", (e: MouseEvent) => {
          const rect = containerRef.current!.getBoundingClientRect();
          const px = e.clientX - rect.left;
          const py = e.clientY - rect.top;
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const coords = (mapRef.current as any).unproject([px, py]);
          if (!coords) return;

          const [lat, lon] = coords;
          onChange({ lat, lon });

          if (markerRef.current) {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (mapRef.current as any).removeChild(markerRef.current);
          }
          const el = createMarkerElement();
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const marker = new (ymaps3 as any).YMapMarker({ coordinates: [lat, lon] }, el);
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          (mapRef.current as any).addChild(marker);
          markerRef.current = marker;
        });

        setLoaded(true);
      })
      .catch((err) => {
        if (!cancelled) setError(err.message);
      });

    return () => {
      cancelled = true;
      if (mapRef.current) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (mapRef.current as any).destroy?.();
        mapRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [resolvedApiKey]);

  // Keep marker in sync when value changes externally
  useEffect(() => {
    if (!loaded || !mapRef.current) return;
    const ymaps3 = window.ymaps3;
    if (!ymaps3) return;

    if (markerRef.current) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (mapRef.current as any).removeChild(markerRef.current);
      markerRef.current = null;
    }

    if (value) {
      const el = createMarkerElement();
      const marker = new ymaps3.YMapMarker({ coordinates: [value.lat, value.lon] }, el);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (mapRef.current as any).addChild(marker);
      markerRef.current = marker;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (mapRef.current as any).setLocation({ center: [value.lat, value.lon], zoom: 14, duration: 300 });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value?.lat, value?.lon, loaded]);

  if (error) {
    return (
      <div style={{ height, display: "flex", alignItems: "center", justifyContent: "center", border: "1px solid #ccc", borderRadius: 8, color: "red", fontSize: 14 }}>
        {error}
      </div>
    );
  }

  return (
    <div style={{ position: "relative" }}>
      <div ref={containerRef} style={{ height, borderRadius: 8, overflow: "hidden", border: "1px solid #ccc" }} />
      {value && (
        <button
          type="button"
          onClick={() => onChange(null)}
          style={{
            position: "absolute",
            top: 8,
            right: 8,
            background: "white",
            border: "1px solid #ccc",
            borderRadius: 4,
            padding: "4px 8px",
            cursor: "pointer",
            fontSize: 12,
          }}
        >
          Убрать точку
        </button>
      )}
      {!loaded && !error && (
        <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "center", justifyContent: "center", background: "#f5f5f5", borderRadius: 8 }}>
          Загрузка карты…
        </div>
      )}
    </div>
  );
}

function createMarkerElement(): HTMLElement {
  const el = document.createElement("div");
  el.style.cssText = "width:20px;height:20px;background:#e74c3c;border-radius:50%;border:3px solid white;box-shadow:0 2px 6px rgba(0,0,0,0.4);cursor:pointer;transform:translate(-50%,-50%)";
  return el;
}
