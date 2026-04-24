import { useEffect, useRef, useState } from "react";
import { DEFAULT_CENTER, loadYmaps, type LatLon } from "@/shared/ui/yandexMaps";
export type { LatLon } from "@/shared/ui/yandexMaps";

interface Props {
  value: LatLon | null;
  onChange: (value: LatLon | null) => void;
  height?: string;
  /** Optional API key. Falls back to VITE_YANDEX_MAPS_API_KEY env variable. */
  apiKey?: string;
}

export default function YandexMapPicker({ value, onChange, height = "300px", apiKey }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<unknown>(null);
  const markerRef = useRef<unknown>(null);
  const [runtimeError, setRuntimeError] = useState<string | null>(null);
  const [loaded, setLoaded] = useState(false);

  const resolvedApiKey = apiKey ?? import.meta.env.VITE_YANDEX_MAPS_API_KEY ?? "";
  const configError = resolvedApiKey
    ? null
    : "Яндекс Карты: API-ключ не задан (VITE_YANDEX_MAPS_API_KEY)";
  const error = configError ?? runtimeError;

  useEffect(() => {
    if (configError) {
      return;
    }

    let cancelled = false;

    loadYmaps(resolvedApiKey)
      .then(() => {
        if (cancelled || !containerRef.current) return;

        const ymaps = window.ymaps;
        const center: [number, number] = value ? [value.lat, value.lon] : DEFAULT_CENTER;

        const map = new ymaps.Map(containerRef.current, {
          center,
          zoom: value ? 14 : 10,
          controls: ["zoomControl", "fullscreenControl"],
        });

        mapRef.current = map;

        if (value) {
          const placemark = new ymaps.Placemark([value.lat, value.lon], {}, { preset: "islands#redDotIcon" });
          map.geoObjects.add(placemark);
          markerRef.current = placemark;
        }

        map.events.add("click", (e: unknown) => {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const coords: [number, number] = (e as any).get("coords");
          const [lat, lon] = coords;
          onChange({ lat, lon });

          if (markerRef.current) {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            (mapRef.current as any).geoObjects.remove(markerRef.current);
          }
          const placemark = new ymaps.Placemark([lat, lon], {}, { preset: "islands#redDotIcon" });
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          (mapRef.current as any).geoObjects.add(placemark);
          markerRef.current = placemark;
        });

        setLoaded(true);
      })
      .catch((err: Error) => {
        if (!cancelled) setRuntimeError(err.message);
      });

    return () => {
      cancelled = true;
      if (mapRef.current) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (mapRef.current as any).destroy();
        mapRef.current = null;
        markerRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [configError, resolvedApiKey]);

  // Sync marker when value changes externally
  useEffect(() => {
    if (!loaded || !mapRef.current) return;
    const ymaps = window.ymaps;
    if (!ymaps) return;

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const map = mapRef.current as any;

    if (markerRef.current) {
      map.geoObjects.remove(markerRef.current);
      markerRef.current = null;
    }

    if (value) {
      const placemark = new ymaps.Placemark([value.lat, value.lon], {}, { preset: "islands#redDotIcon" });
      map.geoObjects.add(placemark);
      markerRef.current = placemark;
      map.setCenter([value.lat, value.lon], 14, { duration: 300 });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value?.lat, value?.lon, loaded]);

  if (error) {
    return (
      <div style={{ height, display: "flex", alignItems: "center", justifyContent: "center", border: "1px solid #ccc", borderRadius: 8, color: "red", fontSize: 14, padding: 16, textAlign: "center" }}>
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
