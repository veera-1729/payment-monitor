import React, { useState, useEffect } from 'react';
import { Box, Container, Grid, Paper, Typography } from '@mui/material';
import MetricsChart from '../components/MetricsChart';
import DimensionSelector from '../components/DimensionSelector';
import AlertsList from '../components/AlertsList';

const DIMENSIONS = ['gateway', 'gateway_payment_method', 'gateway_merchant'];

function Dashboard() {
  const [selectedDimension, setSelectedDimension] = useState('gateway');
  const [metricsData, setMetricsData] = useState([]);
  const [alerts, setAlerts] = useState([]);

  useEffect(() => {
    // Set up WebSocket connection for real-time updates
    const ws = new WebSocket('ws://localhost:8080/ws');

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      console.log('WebSocket message received:', data);
      if (data.type === 'metrics') {
        setMetricsData(prevData => {
          const newData = [...prevData, data];
          // Keep only last 60 data points (1 minute of data at 1 second interval)
          const filteredData = newData.slice(-60);
          console.log('Updated metrics data:', filteredData);
          return filteredData;
        });
      } else if (data.type === 'alert') {
        console.log('Alert received:', data);
        setAlerts(prevAlerts => [data, ...prevAlerts].slice(0, 10));
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    return () => {
      ws.close();
    };
  }, []);

  return (
    <Box component="main" sx={{ flexGrow: 1, py: 4 }}>
      <Container maxWidth="xl">
        <Typography variant="h4" sx={{ mb: 4 }}>
          Payment Success Rate Monitor
        </Typography>
        
        <Grid container spacing={3}>
          {/* Dimension Selector */}
          <Grid item xs={12}>
            <Paper sx={{ p: 2 }}>
              <DimensionSelector
                dimensions={DIMENSIONS}
                selectedDimension={selectedDimension}
                onDimensionChange={setSelectedDimension}
              />
            </Paper>
          </Grid>

          {/* Metrics Chart */}
          <Grid item xs={12} lg={8}>
            <Paper sx={{ p: 2, height: '400px' }}>
              <MetricsChart
                data={metricsData.filter(d => d.dimension === selectedDimension)}
                dimension={selectedDimension}
              />
            </Paper>
          </Grid>

          {/* Alerts List */}
          <Grid item xs={12} lg={4}>
            <Paper sx={{ p: 2, height: '400px', overflow: 'auto' }}>
              <AlertsList alerts={alerts} />
            </Paper>
          </Grid>
        </Grid>
      </Container>
    </Box>
  );
}

export default Dashboard; 