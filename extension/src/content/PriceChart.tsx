import React, { useEffect, useState } from 'react';
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Legend
} from 'recharts';
import moment from 'moment-jalaali';

interface PriceChartProps {
    dkpId: string;
    variantId: string | null;
}

interface VariantSeries {
    variant_id: string;
    columns: string[];
    data: any[][];
}

interface PriceData {
    dkp_id: string;
    columns: string[];
    data: any[][];
    variants?: VariantSeries[];
}

interface ChartDataPoint {
    time: number;
    date: string;
    price: number;
    seller_id: string;
    is_buy_box: boolean;
    variant_id?: string;
}

const PriceChart: React.FC<PriceChartProps> = ({ dkpId, variantId }) => {
    const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isExpanded, setIsExpanded] = useState(true);

    useEffect(() => {
        fetchPriceHistory();
    }, [dkpId, variantId]);

    // Format date to Jalali (Shamsi) calendar
    const formatDate = (timestamp: number) => {
        // Handle invalid timestamps (negative or zero)
        if (!timestamp || timestamp <= 0) {
            return 'تاریخ نامعتبر';
        }
        return moment(timestamp * 1000).format('jYYYY/jMM/jDD');
    };

    // Format date for tooltip with day name in Persian
    const formatTooltipDate = (timestamp: number) => {
        // Handle invalid timestamps
        if (!timestamp || timestamp <= 0) {
            return 'تاریخ نامعتبر';
        }
        moment.loadPersian({ dialect: 'persian-modern', usePersianDigits: true });
        const m = moment(timestamp * 1000);
        return m.format('jYYYY/jMM/jDD - dddd');
    };

    const fetchPriceHistory = async () => {
        try {
            setLoading(true);
            setError(null);

            const response = await chrome.runtime.sendMessage({
                type: 'GET_PRICE_HISTORY',
                dkpId,
                variantId
            });

            if (response.success) {
                const data: PriceData = response.data;

                // If we have variants array (new format), merge all variant data
                if (data.variants && data.variants.length > 0) {
                    // Merge all variants into a single timeline with variant_id
                    const allData: any[] = [];
                    data.variants.forEach(variant => {
                        variant.data
                            .filter((row) => row[0] > 0)
                            .forEach(row => {
                                allData.push({
                                    time: row[0],
                                    date: formatDate(row[0]),
                                    price: row[1],
                                    seller_id: row[2],
                                    is_buy_box: row[3],
                                    variant_id: variant.variant_id,
                                });
                            });
                    });
                    
                    // Sort by time
                    allData.sort((a, b) => a.time - b.time);
                    setChartData(allData);
                } else {
                    // Fallback to old format for backward compatibility
                    const transformed = data.data
                        .filter((row) => row[0] > 0)
                        .map((row) => ({
                            time: row[0],
                            date: formatDate(row[0]),
                            price: row[1],
                            seller_id: row[2],
                            is_buy_box: row[3],
                            variant_id: row[4] || 'unknown',
                        }));

                    setChartData(transformed);
                }
            } else {
                setError(response.error || 'Failed to fetch price history');
            }
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        } finally {
            setLoading(false);
        }
    };

    const formatPrice = (value: number) => {
        return `${(value / 1000).toLocaleString('fa-IR')} K`;
    };

    const formatFullPrice = (price: number) => {
        return new Intl.NumberFormat('fa-IR').format(price);
    };

    const CustomTooltip = ({ active, payload }: any) => {
        if (active && payload && payload.length) {
            const data = payload[0].payload;
            return (
                <div
                    style={{
                        background: 'white',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                        padding: '10px',
                        boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
                    }}
                >
                    <p style={{ margin: '0 0 8px 0', fontWeight: 'bold', fontSize: '14px' }}>
                        {formatTooltipDate(data.time)}
                    </p>
                    {data.variant_id && (
                        <p style={{ margin: '4px 0', fontSize: '12px', color: '#9C27B0', fontWeight: 'bold' }}>
                            🏷️ Variant: {data.variant_id}
                        </p>
                    )}
                    <p style={{ margin: '4px 0', color: '#1890ff', fontSize: '13px' }}>
                        قیمت: {formatFullPrice(data.price)} تومان
                    </p>
                    {data.seller_id && (
                        <p style={{ margin: '4px 0', fontSize: '12px', color: '#666' }}>
                            فروشنده: {data.seller_id}
                        </p>
                    )}
                    {data.is_buy_box && (
                        <p style={{ margin: '4px 0', fontSize: '11px', color: '#52c41a' }}>
                            ✓ Buy Box
                        </p>
                    )}
                </div>
            );
        }
        return null;
    };

    if (loading) {
        return (
            <div style={styles.container}>
                <div style={styles.header}>
                    <h3 style={styles.title}>📊 تاریخچه قیمت - Keepa</h3>
                </div>
                <div style={styles.loading}>در حال بارگذاری...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div style={styles.container}>
                <div style={styles.header}>
                    <h3 style={styles.title}>📊 تاریخچه قیمت - Keepa</h3>
                </div>
                <div style={styles.error}>
                    خطا: {error}
                    <button style={styles.retryButton} onClick={fetchPriceHistory}>
                        تلاش مجدد
                    </button>
                </div>
            </div>
        );
    }

    if (chartData.length === 0) {
        return (
            <div style={styles.container}>
                <div style={styles.header}>
                    <h3 style={styles.title}>📊 تاریخچه قیمت - Keepa</h3>
                </div>
                <div style={styles.noData}>
                    هنوز داده‌ای برای این محصول ثبت نشده است.
                </div>
            </div>
        );
    }

    const minPrice = Math.min(...chartData.map(d => d.price));
    const maxPrice = Math.max(...chartData.map(d => d.price));
    const avgPrice = chartData.reduce((sum, d) => sum + d.price, 0) / chartData.length;

    return (
        <div style={styles.container}>
            <div
                style={styles.header}
                onClick={() => setIsExpanded(!isExpanded)}
            >
                <h3 style={styles.title}>📊 تاریخچه قیمت - Keepa</h3>
                <button style={styles.toggleButton}>
                    {isExpanded ? '▼' : '▲'}
                </button>
            </div>

            {isExpanded && (
                <>
                    <div style={styles.stats}>
                        <div style={styles.stat}>
                            <span style={styles.statLabel}>کمترین:</span>
                            <span style={styles.statValue}>{minPrice.toLocaleString('fa-IR')}</span>
                        </div>
                        <div style={styles.stat}>
                            <span style={styles.statLabel}>میانگین:</span>
                            <span style={styles.statValue}>{Math.round(avgPrice).toLocaleString('fa-IR')}</span>
                        </div>
                        <div style={styles.stat}>
                            <span style={styles.statLabel}>بیشترین:</span>
                            <span style={styles.statValue}>{maxPrice.toLocaleString('fa-IR')}</span>
                        </div>
                    </div>

                    <div style={styles.chartContainer}>
                        <ResponsiveContainer width="100%" height={300}>
                            <LineChart data={chartData}>
                                <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
                                <XAxis
                                    dataKey="date"
                                    tick={{ fontSize: 12 }}
                                    stroke="#666"
                                />
                                <YAxis
                                    tickFormatter={formatPrice}
                                    tick={{ fontSize: 12 }}
                                    stroke="#666"
                                />
                                <Tooltip content={<CustomTooltip />} />
                                <Legend />
                                {(() => {
                                    // Get unique variant IDs
                                    const variantIds = Array.from(new Set(chartData.map(d => d.variant_id))).filter(Boolean);
                                    
                                    // Color palette for up to 25 variants
                                    const colors = [
                                        '#4CAF50', '#2196F3', '#FF5722', '#9C27B0', '#FF9800',
                                        '#00BCD4', '#E91E63', '#8BC34A', '#FFC107', '#3F51B5',
                                        '#009688', '#F44336', '#CDDC39', '#673AB7', '#FF6F00',
                                        '#00897B', '#C2185B', '#7CB342', '#FFA000', '#5E35B1',
                                        '#00ACC1', '#D32F2F', '#AFB42B', '#512DA8', '#FF8F00'
                                    ];
                                    
                                    // If only one variant or no variant info, show single line
                                    if (variantIds.length <= 1) {
                                        return (
                                            <Line
                                                type="monotone"
                                                dataKey="price"
                                                stroke="#4CAF50"
                                                strokeWidth={2}
                                                dot={{ fill: '#4CAF50', r: 4 }}
                                                activeDot={{ r: 6 }}
                                                name="قیمت (تومان)"
                                            />
                                        );
                                    }
                                    
                                    // Create a line for each variant
                                    return variantIds.map((variantId, index) => {
                                        const color = colors[index % colors.length];
                                        // Filter data for this variant
                                        const variantData = chartData.filter(d => d.variant_id === variantId);
                                        
                                        return (
                                            <Line
                                                key={variantId}
                                                type="monotone"
                                                dataKey="price"
                                                data={variantData}
                                                stroke={color}
                                                strokeWidth={2}
                                                dot={{ fill: color, r: 3 }}
                                                activeDot={{ r: 5 }}
                                                name={`Variant ${variantId}`}
                                                connectNulls
                                            />
                                        );
                                    });
                                })()}
                            </LineChart>
                        </ResponsiveContainer>
                    </div>
                </>
            )}
        </div>
    );
};

const styles: { [key: string]: React.CSSProperties } = {
    container: {
        backgroundColor: '#ffffff',
        border: '2px solid #e0e0e0',
        borderRadius: '12px',
        boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
        margin: '20px 0',
        fontFamily: 'Arial, sans-serif',
        overflow: 'hidden',
    },
    header: {
        background: 'linear-gradient(135deg, #4CAF50 0%, #45a049 100%)',
        color: 'white',
        padding: '16px 20px',
        cursor: 'pointer',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
    },
    title: {
        margin: 0,
        fontSize: '18px',
        fontWeight: 'bold',
    },
    toggleButton: {
        background: 'transparent',
        border: 'none',
        color: 'white',
        fontSize: '18px',
        cursor: 'pointer',
        padding: '4px 8px',
    },
    stats: {
        display: 'flex',
        justifyContent: 'space-around',
        padding: '16px',
        backgroundColor: '#f5f5f5',
        borderBottom: '1px solid #e0e0e0',
    },
    stat: {
        textAlign: 'center',
    },
    statLabel: {
        display: 'block',
        fontSize: '12px',
        color: '#666',
        marginBottom: '4px',
    },
    statValue: {
        display: 'block',
        fontSize: '16px',
        fontWeight: 'bold',
        color: '#333',
    },
    chartContainer: {
        padding: '20px',
    },
    loading: {
        padding: '40px',
        textAlign: 'center',
        color: '#666',
        fontSize: '16px',
    },
    error: {
        padding: '40px',
        textAlign: 'center',
        color: '#f44336',
        fontSize: '14px',
    },
    retryButton: {
        marginTop: '12px',
        padding: '8px 16px',
        backgroundColor: '#4CAF50',
        color: 'white',
        border: 'none',
        borderRadius: '4px',
        cursor: 'pointer',
        fontSize: '14px',
    },
    noData: {
        padding: '40px',
        textAlign: 'center',
        color: '#666',
        fontSize: '14px',
    },
};

export default PriceChart;
