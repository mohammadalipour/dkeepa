import React, { useEffect, useRef } from 'react';
import ReactDOM from 'react-dom/client';

interface ShadowRootProps {
    children: React.ReactNode;
}

/**
 * ShadowRoot component that renders children inside a Shadow DOM
 * This provides complete CSS isolation from the host page
 */
const ShadowRoot: React.FC<ShadowRootProps> = ({ children }) => {
    const hostRef = useRef<HTMLDivElement>(null);
    const shadowRootRef = useRef<ShadowRoot | null>(null);
    const reactRootRef = useRef<ReactDOM.Root | null>(null);

    useEffect(() => {
        if (!hostRef.current) return;

        // Create shadow root if it doesn't exist
        if (!shadowRootRef.current) {
            shadowRootRef.current = hostRef.current.attachShadow({ mode: 'open' });

            // Inject styles into shadow DOM
            const style = document.createElement('style');
            style.textContent = `
        * {
          box-sizing: border-box;
          margin: 0;
          padding: 0;
        }
        
        :host {
          all: initial;
          display: block;
        }
      `;
            shadowRootRef.current.appendChild(style);

            // Create container for React app
            const container = document.createElement('div');
            container.id = 'keepa-shadow-root';
            shadowRootRef.current.appendChild(container);

            // Create React root
            reactRootRef.current = ReactDOM.createRoot(container);
        }

        // Render children
        if (reactRootRef.current) {
            reactRootRef.current.render(children);
        }

        // Cleanup
        return () => {
            if (reactRootRef.current) {
                reactRootRef.current.unmount();
                reactRootRef.current = null;
            }
        };
    }, [children]);

    return <div ref={hostRef} />;
};

export default ShadowRoot;
