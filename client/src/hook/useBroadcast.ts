import { useEffect } from "react";
import { connect, StringCodec, type NatsConnection, type Subscription } from "nats.ws";

export function useBroadcast<T = any>(
  subject: string,
  onMessage: (message: T) => void,
  onError: (error: Error) => void
): void {
  useEffect(() => {
    let nc: NatsConnection;
    let sub: Subscription;
    const sc = StringCodec();

    (async () => {
      try {
        
        nc = await connect({ servers: import.meta.env.VITE_BROADCAST_URL });
        
        sub = nc.subscribe(subject);
        console.log(`💬 Subscribed to ${subject}`);
        for await (const m of sub) {
          try {
            const decoded = sc.decode(m.data);
            const parsed = JSON.parse(decoded) as T;
            onMessage(parsed); // Trigger the callback on receiving a message
          } catch (parseErr) {
            console.error("Failed to parse NATS message:", parseErr);
          }
        }
      } catch (err: any) {
        onError(err); // Trigger the callback if there's an error
      }
    })();

    return () => {
      if (sub) sub.unsubscribe();
      if (nc) nc.close();
    };
  }, [subject, onMessage, onError]);
}
