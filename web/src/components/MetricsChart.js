import React from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { format } from 'date-fns';
import { Typography, Box } from '@mui/material';

const COLORS = [
  '#8884d8',
  '#82ca9d',
  '#ffc658',
  '#ff7300',
  '#00C49F',
  '#FFBB28',
  '#FF8042',
];

function MetricsChart({ data, dimension }) {
  // Group data by value (e.g., different gateways)
  const groupedData = data.reduce((acc, item) => {
    const timestamp = format(new Date(item.timestamp), 'HH:mm:ss');
    if (!acc[timestamp]) {
      acc[timestamp] = { timestamp };
    }
    acc[timestamp][item.value] = item.success_rate;
    return acc;
  }, {});

  const chartData = Object.values(groupedData).sort((a, b) => {
    const timeA = new Date(`1970/01/01 ${a.timestamp}`);
    const timeB = new Date(`1970/01/01 ${b.timestamp}`);
    return timeA - timeB;
  });
  
  // Get unique values (e.g., gateway names)
  const values = [...new Set(data.map(item => item.value))];

  if (data.length === 0) {
    return (
      <Box sx={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Typography variant="body1" color="text.secondary">
          No data available for the selected dimension
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ height: '100%', width: '100%' }}>
      <Typography variant="h6" sx={{ mb: 2 }}>
        Success Rate by {dimension.split('_').map(word => 
          word.charAt(0).toUpperCase() + word.slice(1)
        ).join(' ')}
      </Typography>
      <ResponsiveContainer width="100%" height="85%">
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            dataKey="timestamp"
            tick={{ fontSize: 12 }}
            interval={Math.floor(chartData.length / 10)}
          />
          <YAxis
            domain={[0, 100]}
            tick={{ fontSize: 12 }}
            label={{ 
              value: 'Success Rate (%)',
              angle: -90,
              position: 'insideLeft',
              style: { textAnchor: 'middle' }
            }}
          />
          <Tooltip
            formatter={(value) => [`${value.toFixed(2)}%`, 'Success Rate']}
            labelStyle={{ color: '#000' }}
          />
          <Legend />
          {values.map((value, index) => (
            <Line
              key={value}
              type="monotone"
              dataKey={value}
              name={value}
              stroke={COLORS[index % COLORS.length]}
              strokeWidth={2}
              dot={false}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </Box>
  );
}

export default MetricsChart; 