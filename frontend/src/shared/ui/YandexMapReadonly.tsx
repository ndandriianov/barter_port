import { useEffect, useRef, useState } from "react";
import { loadYmaps, type LatLon } from "@/shared/ui/yandexMaps";

interface Props {
  value: LatLon;
  height?: string;
  /** Optional API key. Falls back to VITE_YANDEX_MAPS_API_KEY env variable. */
  apiKey?: string;
}

export default function YandexMapReadonly({ value, height = "300px", apiKey }: Props) {
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
        const map = new ymaps.Map(containerRef.current, {
          center: [value.lat, value.lon],
          zoom: 14,
          controls: ["zoomControl", "fullscreenControl"],
        });

        const placemark = new ymaps.Placemark([value.lat, value.lon], {}, { preset: "islands#redDotIcon" });
        map.geoObjects.add(placemark);

        mapRef.current = map;
        markerRef.current = placemark;
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
  }, [configError, resolvedApiKey]);

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

    const placemark = new ymaps.Placemark([value.lat, value.lon], {}, { preset: "islands#redDotIcon" });
    map.geoObjects.add(placemark);
    markerRef.current = placemark;
    map.setCenter([value.lat, value.lon], 14, { duration: 300 });
  }, [loaded, value.lat, value.lon]);

  if (error) {
    return (
      <div style={{ height, display: "flex", alignItems: "center", justifyContent: "center", border: "1px solid #ccc", borderRadius: 8, color: "red", fontSize: 14, padding: 16, textAlign: "center" }}>
        {error}
      </div>
    );
  }

  return (
    <div style={{ position: "relative" }}>
      <div ref={containerRef} style={{ height, borderRadius: 8, overflow: "hidden", border: "1px solid #ccc", background: "#f5f5f5" }} />
      {!loaded && !error && (
        <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "center", justifyContent: "center", background: "#f5f5f5", borderRadius: 8 }}>
          Загрузка карты…
        </div>
      )}
    </div>
  );
}
