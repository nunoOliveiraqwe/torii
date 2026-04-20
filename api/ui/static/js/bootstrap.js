// ---------- Bootstrap ----------
initChart();
connectMetricsSSE();
startChartTick();
loadProxyRoutes();
loadIdentity();
loadRecentBlocked();  // seed threat defense cards on dashboard
setInterval(loadProxyRoutes, 10000);
setInterval(tickSparklines, 1000);

