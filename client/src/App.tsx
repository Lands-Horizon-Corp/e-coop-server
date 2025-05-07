"use client"

import { useEffect, useRef } from "react";

function App() {
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {

    const socket = new WebSocket(import.meta.env.VITE_BROADCAST_URL);
    socketRef.current = socket;

    socket.onopen = () => {
      console.log('WebSocket connected');

      // Example: subscribe to a topic
      socket.send(JSON.stringify({ topic: 'chat.general' }));
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log('Received:', data);
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    socket.onclose = () => {
      console.log('WebSocket closed');
    };

    return () => {
      socket.close();
    };
  }, []);
    return (
      <div>
        <div>
          {/* <SampleMedia/> */}
        </div>
        <div>
          {/* <SampleForm/> */}
        </div>
      </div>
    )
}

export default App
