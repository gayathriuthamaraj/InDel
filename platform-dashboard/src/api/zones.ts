import { coreClient } from './client';

// Get all zones (optionally by level)
export const getZones = (level?: 'a' | 'b' | 'c' | 'A' | 'B' | 'C') =>
  coreClient.get('/api/v1/platform/zones', { params: level ? { level: level.toUpperCase() } : undefined });

// Get zone health
export const getZoneHealth = () => coreClient.get('/api/v1/platform/zones/health');

// Get zone paths for a given type (a, b, c)
export const getZonePaths = (type: 'a' | 'b' | 'c') =>
  coreClient.get(`/api/v1/platform/zone-paths?type=${type}`);
