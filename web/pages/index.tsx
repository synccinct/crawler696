// web/pages/index.tsx
import React, { useState, useEffect } from 'react';
import { 
  Container, 
  Typography, 
  Grid, 
  Paper, 
  Button,
  Card,
  CardContent,
  CardHeader,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  LinearProgress,
  Chip
} from '@mui/material';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import axios from 'axios';

interface CrawlSession {
  id: string;
  name: string;
  description: string;
  status: string;
  created_at: string;
  stats: {
    total_tasks: number;
    completed_tasks: number;
    failed_tasks: number;
    pending_tasks: number;
    pages_per_minute: number;
  };
}

interface SystemStats {
  crawler: {
    TotalRequests: number;
    SuccessfulCrawls: number;
    FailedCrawls: number;
    DetectionEvents: number;
    ActiveWorkers: number;
    QueueSize: number;
  };
  proxies: Record<string, any>;
  timestamp: string;
}

export default function Dashboard() {
  const [sessions, setSessions] = useState<CrawlSession[]>([]);
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000); // Refresh every 5 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchData = async () => {
    try {
      const [sessionsRes, statsRes] = await Promise.all([
        axios.get('/api/v1/crawls'),
        axios.get('/api/v1/stats')
      ]);
      
      setSessions(sessionsRes.data);
      setStats(statsRes.data);
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch data:', error);
      setLoading(false);
    }
  };

  const getSuccessRate = () => {
    if (!stats?.crawler) return 0;
    const { SuccessfulCrawls, TotalRequests } = stats.crawler;
    return TotalRequests > 0 ? (SuccessfulCrawls / TotalRequests * 100).toFixed(1) : 0;
  };

  const getDetectionRate = () => {
    if (!stats?.crawler) return 0;
    const { DetectionEvents, TotalRequests } = stats.crawler;
    return TotalRequests > 0 ? (DetectionEvents / TotalRequests * 100).toFixed(2) : 0;
  };

  if (loading) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <LinearProgress />
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h3" component="h1" gutterBottom sx={{ 
        fontWeight: 'bold',
        background: 'linear-gradient(45deg, #FF6B6B, #4ECDC4)',
        WebkitBackgroundClip: 'text',
        WebkitTextFillColor: 'transparent',
        textAlign: 'center',
        mb: 4
      }}>
        🕷️ CRAWLER666 Command Center
      </Typography>

      {/* System Overview Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ backgroundColor: '#1a1a1a', color: 'white' }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Active Workers</Typography>
              <Typography variant="h4" color="primary">
                {stats?.crawler.ActiveWorkers || 0}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ backgroundColor: '#1a1a1a', color: 'white' }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Queue Size</Typography>
              <Typography variant="h4" color="warning.main">
                {stats?.crawler.QueueSize || 0}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        
        <Grid item xs={12} sm={6} md={3}>
          <Card sx={{ backgroundColor: '#1a1a1a', color: 'white' }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Success Rate</Typography>
              <Typography variant="h4" color="success.main">
                {getSuccessRate()}%
              </Typography>
            </CardContent>
          </Card>
      
