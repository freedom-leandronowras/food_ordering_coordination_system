"use client";

import { useState, useEffect } from "react";

export default function Home() {
  const [data, setData] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = () => {
      fetch("/api/health/hello")
        .then((res) => {
          if (!res.ok) {
            throw new Error(`Error: ${res.status} ${res.statusText}`);
          }
          return res.text();
        })
        .then((text) => setData(text))
        .catch((err) => {
          console.error("Fetch failed:", err);
          setError(err.message);
        });
    };

    // Initial fetch
    fetchData();

    // Poll every 2 seconds
    const interval = setInterval(fetchData, 2000);

    // Clean up interval on unmount
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen">
      <h1 className="text-2xl font-bold mb-4">Go on Vercel</h1>
      <div className="p-6 border rounded-lg shadow-md bg-white">
        {error ? (
          <p className="text-red-500 font-semibold">Failed to load: {error}</p>
        ) : (
          <p className="text-gray-800">{data || "Loading..."}</p>
        )}
      </div>
    </div>
  );
}
