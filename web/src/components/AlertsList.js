import React from 'react';
import {
  List,
  ListItem,
  ListItemText,
  Typography,
  Box,
  Chip,
  Divider,
} from '@mui/material';
import { format } from 'date-fns';

function AlertsList({ alerts }) {
  if (alerts.length === 0) {
    return (
      <Box sx={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Typography variant="body1" color="text.secondary">
          No alerts to display
        </Typography>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h6" sx={{ mb: 2 }}>
        Recent Alerts
      </Typography>
      <List>
        {alerts.map((alert, index) => (
          <React.Fragment key={alert.id}>
            <ListItem
              alignItems="flex-start"
              sx={{
                bgcolor: 'background.paper',
                borderRadius: 1,
                mb: 1,
              }}
            >
              <ListItemText
                primary={
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                    <Typography variant="subtitle1" component="span">
                      {alert.dimension.split('_').map(word => 
                        word.charAt(0).toUpperCase() + word.slice(1)
                      ).join(' ')}
                    </Typography>
                    <Chip
                      label={alert.value}
                      size="small"
                      color="primary"
                      sx={{ ml: 1 }}
                    />
                  </Box>
                }
                secondary={
                  <React.Fragment>
                    <Typography
                      component="span"
                      variant="body2"
                      color="text.primary"
                      sx={{ display: 'block', mb: 0.5 }}
                    >
                      Success Rate: {alert.current_rate.toFixed(2)}% (Previous: {alert.previous_rate.toFixed(2)}%)
                    </Typography>
                    <Typography
                      component="span"
                      variant="body2"
                      color="error"
                      sx={{ display: 'block', mb: 0.5 }}
                    >
                      Drop: {alert.drop_percentage.toFixed(2)}%
                    </Typography>
                    {alert.root_cause && (
                      <>
                        <Typography
                          component="span"
                          variant="body2"
                          color="text.primary"
                          sx={{ display: 'block', mt: 1, fontWeight: 'bold' }}
                        >
                          Root Cause: 
                        </Typography>
                        <Typography
                          component="span"
                          variant="body2"
                          color="text.primary"
                          sx={{ display: 'block', mb: 0.5 }}
                        >
                          {alert.root_cause}
                        </Typography>
                      </>
                    )}
                    {alert.confidence > 0 && (
                      <Typography
                        component="span"
                        variant="body2"
                        color="text.secondary"
                        sx={{ display: 'block', mb: 0.5 }}
                      >
                        Confidence: {(alert.confidence * 100).toFixed(0)}%
                      </Typography>
                    )}
                    {alert.recommendations && alert.recommendations.length > 0 && (
                      <>
                        <Typography
                          component="span"
                          variant="body2"
                          color="text.primary"
                          sx={{ display: 'block', mt: 1, fontWeight: 'bold' }}
                        >
                          Recommendations:
                        </Typography>
                        <ul style={{ margin: '4px 0', paddingLeft: '20px' }}>
                          {alert.recommendations.map((rec, i) => (
                            <li key={i}>
                              <Typography
                                component="span"
                                variant="body2"
                                color="text.primary"
                              >
                                {rec}
                              </Typography>
                            </li>
                          ))}
                        </ul>
                      </>
                    )}
                    <Typography
                      component="span"
                      variant="caption"
                      color="text.secondary"
                    >
                      {format(new Date(alert.timestamp), 'MMM d, yyyy HH:mm:ss')}
                    </Typography>
                  </React.Fragment>
                }
              />
            </ListItem>
            {index < alerts.length - 1 && <Divider />}
          </React.Fragment>
        ))}
      </List>
    </Box>
  );
}

export default AlertsList; 