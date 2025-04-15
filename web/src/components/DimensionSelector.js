import React from 'react';
import { Box, ToggleButton, ToggleButtonGroup, Typography } from '@mui/material';

function DimensionSelector({ dimensions, selectedDimension, onDimensionChange }) {
  const handleChange = (event, newDimension) => {
    if (newDimension !== null) {
      onDimensionChange(newDimension);
    }
  };

  return (
    <Box>
      <Typography variant="h6" sx={{ mb: 2 }}>
        Select Dimension
      </Typography>
      <ToggleButtonGroup
        value={selectedDimension}
        exclusive
        onChange={handleChange}
        aria-label="dimension selector"
      >
        {dimensions.map((dimension) => (
          <ToggleButton
            key={dimension}
            value={dimension}
            aria-label={dimension}
            sx={{
              textTransform: 'none',
              px: 3,
              py: 1,
            }}
          >
            {dimension.split('_').map(word => 
              word.charAt(0).toUpperCase() + word.slice(1)
            ).join(' ')}
          </ToggleButton>
        ))}
      </ToggleButtonGroup>
    </Box>
  );
}

export default DimensionSelector; 