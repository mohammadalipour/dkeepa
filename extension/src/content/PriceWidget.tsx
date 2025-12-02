import React, { useEffect, useState } from 'react';

interface PriceWidgetProps {
    dkpId: string;
}

interface PriceData {
    dkp_id: string;
    columns: string[];
    data: any[][];
}

const PriceWidget: React.FC<PriceWidgetProps> = ({ dkpId }) => {
    const [priceData, setPriceData] = useState<PriceData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isMinimized, setIsMinimized] = useState(false);

    useEffect(() => {
        // Request price history from background script
        chrome.runtime.sendMessage(
            { type: 'GET_PRICE_HISTORY', dkpId },
            (response) => {
                setLoading(false);
                if (response.success) {
                    setPriceData(response.data);
                } else {
                    setError(response.error);
                }
            }
        );
    }, [dkpId]);

    if (loading) {
        return (
            <div style={styles.widget}>
                <div style={styles.header}>
                    <span style={styles.title}>📊 Keepa</span>
                </div>
                <div style={styles.content}>Loading...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div style={styles.widget}>
                <div style={styles.header}>
                    <span style={styles.title}>📊 Keepa</span>
                </div>
                <div style={styles.content}>Error: {error}</div>
            </div>
        );
    }

    const hasData = priceData && priceData.data.length > 0;
    const latestPrice = hasData ? priceData.data[priceData.data.length - 1][1] : null;

    return (
        <div style={styles.widget}>
            <div style={styles.header} onClick={() => setIsMinimized(!isMinimized)}>
                <span style={styles.title}>📊 Keepa Price Tracker</span>
                <button style={styles.toggleButton}>{isMinimized ? '▼' : '▲'}</button>
            </div>

            {!isMinimized && (
                <div style={styles.content}>
                    {hasData ? (
                        <>
                            <div style={styles.stat}>
                                <strong>Latest Price:</strong> {latestPrice?.toLocaleString()} تومان
                            </div>
                            <div style={styles.stat}>
                                <strong>Data Points:</strong> {priceData.data.length}
                            </div>
                            <div style={styles.stat}>
                                <strong>Product ID:</strong> {dkpId}
                            </div>
                        </>
                    ) : (
                        <div>No price history available yet.</div>
                    )}
                </div>
            )}
        </div>
    );
};

const styles: { [key: string]: React.CSSProperties } = {
    widget: {
        backgroundColor: '#ffffff',
        border: '2px solid #e0e0e0',
        borderRadius: '8px',
        boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
        width: '300px',
        fontFamily: 'Arial, sans-serif',
        fontSize: '14px',
    },
    header: {
        backgroundColor: '#4CAF50',
        color: 'white',
        padding: '12px',
        borderRadius: '6px 6px 0 0',
        cursor: 'pointer',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
    },
    title: {
        fontWeight: 'bold',
        fontSize: '16px',
    },
    toggleButton: {
        background: 'transparent',
        border: 'none',
        color: 'white',
        fontSize: '16px',
        cursor: 'pointer',
    },
    content: {
        padding: '16px',
    },
    stat: {
        marginBottom: '8px',
        padding: '8px',
        backgroundColor: '#f5f5f5',
        borderRadius: '4px',
    },
};

export default PriceWidget;
