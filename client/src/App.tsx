import { useBroadcast } from './hook/useBroadcast';
interface Payload {
  id: string;
  timestamp: string;
  data: any;
} 

function App() {


  const { messages, error } = useBroadcast<Payload>("updates.topic");

  

  return (
    <div>
      <h2>Live Updates</h2>
      {messages.length === 0 && <p>No messages yet…</p>}
      {error?.message}
      <ul>
        {messages.map((msg, idx) => (
          <li key={idx}>
            <strong>{msg.id}</strong> at {new Date(msg.timestamp).toLocaleTimeString()}:{" "}
            {JSON.stringify(msg.data)}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default App
