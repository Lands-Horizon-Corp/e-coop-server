import { useEffect, useState } from "react";
import { connect, StringCodec, type NatsConnection, type Subscription } from "nats.ws";

export function useBroadcast<T = any>(
  subject: string
): { messages: T[]; error: Error | null } {
  const [messages, setMessages] = useState<T[]>([]);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let nc: NatsConnection;
    let sub: Subscription;
    const sc = StringCodec();

    (async () => {
      try {
        nc = await connect({ servers: process.env.REACT_APP_NATS_WS_URL });
        sub = nc.subscribe(subject);
        console.log(`💬 Subscribed to ${subject}`);
        for await (const m of sub) {
          try {
            const decoded = sc.decode(m.data);
            const parsed = JSON.parse(decoded) as T;
            setMessages((old) => [...old, parsed]);
          } catch (parseErr) {
            console.error("Failed to parse NATS message:", parseErr);
          }
        }
      } catch (err: any) {
        setError(err);
      }
    })();

    return () => {
      if (sub) sub.unsubscribe();
      if (nc) nc.close();
    };
  }, [subject]);

  return { messages, error };
}