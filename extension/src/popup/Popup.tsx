import React, { useEffect, useState } from 'react';

const Popup: React.FC = () => {
    const [currentTab, setCurrentTab] = useState<string>('');

    useEffect(() => {
        // Get current tab URL
        chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
            if (tabs[0]?.url) {
                setCurrentTab(tabs[0].url);
            }
        });
    }, []);

    const isDigikalaProduct = currentTab.includes('digikala.com/product/');

    return (
        <div style={styles.container}>
            <div style={styles.header}>
                <h1 style={styles.title}>📊 Keepa</h1>
                <p style={styles.subtitle}>Digikala Price Tracker</p>
            </div>

            <div style={styles.content}>
                {isDigikalaProduct ? (
                    <div>
                        <p style={styles.successText}>✅ Active on this product page</p>
                        <p style={styles.infoText}>
                            Price tracking widget is displayed on the page.
                        </p>
                    </div>
                ) : (
                    <div>
                        <p style={styles.warningText}>⚠️ Not a product page</p>
                        <p style={styles.infoText}>
                            Navigate to a Digikala product page to see price history.
                        </p>
                    </div>
                )}

                <div style={styles.stats}>
                    <h3 style={styles.statsTitle}>Extension Info</h3>
                    <div style={styles.stat}>
                        <strong>Version:</strong> 1.0.0
                    </div>
                    <div style={styles.stat}>
                        <strong>Status:</strong> Active
                    </div>
                </div>
            </div>
        </div>
    );
};

const styles: { [key: string]: React.CSSProperties } = {
    container: {
        width: '350px',
        minHeight: '300px',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    },
    header: {
        background: 'linear-gradient(135deg, #4CAF50 0%, #45a049 100%)',
        color: 'white',
        padding: '20px',
        textAlign: 'center',
    },
    title: {
        margin: '0 0 8px 0',
        fontSize: '24px',
        fontWeight: 'bold',
    },
    subtitle: {
        margin: 0,
        fontSize: '14px',
        opacity: 0.9,
    },
    content: {
        padding: '20px',
    },
    successText: {
        color: '#4CAF50',
        fontWeight: 'bold',
        marginBottom: '8px',
    },
    warningText: {
        color: '#ff9800',
        fontWeight: 'bold',
        marginBottom: '8px',
    },
    infoText: {
        color: '#666',
        fontSize: '14px',
        lineHeight: '1.5',
    },
    stats: {
        marginTop: '20px',
        padding: '16px',
        backgroundColor: '#f5f5f5',
        borderRadius: '8px',
    },
    statsTitle: {
        margin: '0 0 12px 0',
        fontSize: '16px',
        color: '#333',
    },
    stat: {
        marginBottom: '8px',
        fontSize: '14px',
        color: '#666',
    },
};

export default Popup;
