import React from "react";
import useSWR from "swr";
import axios from "axios";

const fetcher = (url: string) => axios.get(url).then(res => res.data);

export default function App() {
  const { data } = useSWR("/api/v1/metrics/summary", fetcher);

  return (
    <main className="p-8">
      <h1 className="text-3xl font-bold">Crawler666 Dashboard</h1>
      <pre className="mt-6 bg-gray-100 p-4 rounded">
        {JSON.stringify(data, null, 2)}
      </pre>
    </main>
  );
}
