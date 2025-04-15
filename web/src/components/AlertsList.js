import React from 'react';
import {
  List,
  ListItem,
  ListItemText,
  Typography,
  Box,
  Chip,
  Divider,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
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
      <List disablePadding>
        {alerts.map((alert, index) => (
          <React.Fragment key={alert.id}>
            <ListItem
              alignItems="flex-start"
              sx={{
                bgcolor: 'background.paper',
                borderRadius: 1,
                mb: 1,
                p: 0,
                flexDirection: 'column'
              }}
            >
              <ListItemText
                sx={{ px: 2, pt: 2, pb: 1 }}
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

              {(alert.root_cause || alert.recommendations?.length > 0 || alert.related_changes?.length > 0) && (
                <Accordion sx={{ width: '100%', boxShadow: 'none', '&:before': { display: 'none' }, borderTop: '1px solid rgba(0, 0, 0, 0.12)' }}>
                  <AccordionSummary
                    expandIcon={<ExpandMoreIcon />}
                    aria-controls={`panel${index}-content`}
                    id={`panel${index}-header`}
                    sx={{ minHeight: '48px', '& .MuiAccordionSummary-content': { margin: '12px 0' } }}
                  >
                    <Typography variant="body2">Analysis Details</Typography>
                    {alert.confidence > 0 && (
                       <Chip label={`Confidence: ${(alert.confidence * 100).toFixed(0)}%`} size="small" sx={{ ml: 2 }} />
                    )}
                  </AccordionSummary>
                  <AccordionDetails sx={{ pt: 0 }}>
                    {alert.root_cause && (
                      <Box sx={{ mb: 1 }}>
                        <Typography variant="subtitle2">Root Cause:</Typography>
                        <Typography variant="body2">{alert.root_cause}</Typography>
                      </Box>
                    )}
                    {alert.recommendations && alert.recommendations.length > 0 && (
                      <Box sx={{ mb: 1 }}>
                        <Typography variant="subtitle2">Recommendations:</Typography>
                        <List dense disablePadding sx={{ pl: 2 }}>
                          {alert.recommendations.map((rec, i) => (
                            <ListItem key={i} disableGutters sx={{ p: 0 }}>
                              <Typography variant="body2">- {rec}</Typography>
                            </ListItem>
                          ))}
                        </List>
                      </Box>
                    )}
                    {alert.related_changes && alert.related_changes.length > 0 && (
                      <Box>
                        <Typography variant="subtitle2">Related Changes:</Typography>
                        <List dense disablePadding sx={{ pl: 2 }}>
                          {alert.related_changes.map((change, i) => (
                            <ListItem key={i} disableGutters sx={{ p: 0 }}>
                              <Typography variant="body2">- {change}</Typography>
                            </ListItem>
                          ))}
                        </List>
                      </Box>
                    )}
                  </AccordionDetails>
                </Accordion>
              )}
            </ListItem>
          </React.Fragment>
        ))}
      </List>
    </Box>
  );
}

export default AlertsList; 