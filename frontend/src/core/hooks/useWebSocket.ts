import { useEffect, useRef, useState, useCallback } from 'react';
import type { WebSocketMessage } from '@/core/types';

interface UseWebSocketOptions {
  mcuCode?: string;
  onMessage?: (message: WebSocketMessage<unknown>) => void;
  onError?: (error: Event) => void;
  reconnectInterval?: number;
}

export const useWebSocket = ({
  mcuCode,
  onMessage,
  onError,
  reconnectInterval = 3000,
}: UseWebSocketOptions) => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WebSocketMessage<unknown> | null>(null);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
  const reconnectAttemptsRef = useRef(0);
  const onMessageRef = useRef(onMessage);
  const onErrorRef = useRef(onError);

  // Update refs when callbacks change
  useEffect(() => {
    onMessageRef.current = onMessage;
    onErrorRef.current = onError;
  }, [onMessage, onError]);

  const sendMessage = useCallback((message: WebSocketMessage<unknown>) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      console.log('📤 Sending WebSocket message:', message);
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.error('❌ WebSocket is not connected. ReadyState:', wsRef.current?.readyState);
      const readyStates = ['CONNECTING', 'OPEN', 'CLOSING', 'CLOSED'];
      console.error('Current state:', readyStates[wsRef.current?.readyState ?? 3]);
    }
  }, []);

  // Main connection effect
  useEffect(() => {
    // Don't connect if no mcuCode
    if (!mcuCode) {
      console.log('⏳ WebSocket: Waiting for MCU code...');
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setConnectionError('Waiting for MCU code');
      setIsConnected(false);
      return;
    }

    const token = localStorage.getItem('token');
    if (!token) {
      console.error('❌ WebSocket: No token found');
      setConnectionError('No authentication token');
      setIsConnected(false);
      return;
    }

    let ws: WebSocket | null = null;
    let reconnectTimeout: NodeJS.Timeout | undefined = undefined;

    const connect = () => {
      try {
        const wsUrl = `ws://localhost:8080/ws?token=${token}&mcu_code=${mcuCode}`;
        console.log('🔌 WebSocket: Connecting to', wsUrl);
        
        ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          console.log('✅ WebSocket connected');
          setIsConnected(true);
          setConnectionError(null);
          reconnectAttemptsRef.current = 0;
        };

        ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage<unknown> = JSON.parse(event.data);
            console.log('📨 Received:', message);
            setLastMessage(message);
            onMessageRef.current?.(message);
          } catch (error) {
            console.error('❌ Parse error:', error);
          }
        };

        ws.onerror = (error) => {
          console.error('❌ WebSocket error:', error);
          setConnectionError('Connection error');
          onErrorRef.current?.(error);
        };

        ws.onclose = (event) => {
          console.log('🔌 Disconnected:', event.code, event.reason);
          setIsConnected(false);
          
          // Auto reconnect with exponential backoff
          reconnectAttemptsRef.current += 1;
          const delay = Math.min(reconnectInterval * reconnectAttemptsRef.current, 30000);
          
          console.log(`🔄 Reconnect in ${delay}ms (attempt ${reconnectAttemptsRef.current})`);
          setConnectionError(`Reconnecting (${reconnectAttemptsRef.current})...`);
          
          reconnectTimeout = setTimeout(() => {
            connect();
          }, delay);
        };
      } catch (error) {
        console.error('❌ Failed to create WebSocket:', error);
        setConnectionError('Connection failed');
      }
    };

    // Initial connection
    connect();

    // Cleanup function
    return () => {
      console.log('🧹 Cleaning up WebSocket...');
      
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
      
      if (ws) {
        ws.onclose = null; // Prevent reconnection on intentional close
        ws.close();
      }
      
      wsRef.current = null;
      setIsConnected(false);
    };
  }, [mcuCode, reconnectInterval]); // Only depend on mcuCode and reconnectInterval

  const disconnect = useCallback(() => {
    console.log('🔌 Manual disconnect requested');
    
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    
    if (wsRef.current) {
      wsRef.current.onclose = null; // Prevent auto-reconnect
      wsRef.current.close();
      wsRef.current = null;
    }
    
    setIsConnected(false);
  }, []);

  const reconnect = useCallback(() => {
    console.log('🔄 Manual reconnect requested');
    disconnect();
    // The effect will reconnect automatically when mcuCode is available
  }, [disconnect]);

  return {
    isConnected,
    lastMessage,
    sendMessage,
    disconnect,
    reconnect,
    connectionError,
  };
};
