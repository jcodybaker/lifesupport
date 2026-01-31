const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

let authToken = localStorage.getItem('auth_token');

export function setAuthToken(token) {
  authToken = token;
  if (token) {
    localStorage.setItem('auth_token', token);
  } else {
    localStorage.removeItem('auth_token');
  }
}

export function getAuthToken() {
  return authToken;
}

export function isAuthenticated() {
  return !!authToken;
}

export function logout() {
  setAuthToken(null);
}

async function fetchAPI(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

// Auth
export async function login(username, password) {
  const data = await fetchAPI('/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
  setAuthToken(data.token);
  return data;
}

// System Status
export async function getSystemStatus() {
  return fetchAPI('/status');
}

// Devices
export async function getDevices() {
  return fetchAPI('/devices');
}

export async function controlDevice(deviceId, action, value = null) {
  return fetchAPI(`/admin/devices/${deviceId}/control`, {
    method: 'POST',
    body: JSON.stringify({ device_id: deviceId, action, value }),
  });
}

export async function createDevice(device) {
  return fetchAPI('/admin/devices', {
    method: 'POST',
    body: JSON.stringify(device),
  });
}

export async function updateDevice(deviceId, device) {
  return fetchAPI(`/admin/devices/${deviceId}`, {
    method: 'PUT',
    body: JSON.stringify(device),
  });
}

export async function deleteDevice(deviceId) {
  return fetchAPI(`/admin/devices/${deviceId}`, {
    method: 'DELETE',
  });
}

// Sensors
export async function getSensors() {
  return fetchAPI('/sensors');
}

export async function getSensorReadings(sensorId, hours = 24) {
  return fetchAPI(`/sensors/${sensorId}/readings?hours=${hours}`);
}

export async function createSensor(sensor) {
  return fetchAPI('/admin/sensors', {
    method: 'POST',
    body: JSON.stringify(sensor),
  });
}

export async function updateSensor(sensorId, sensor) {
  return fetchAPI(`/admin/sensors/${sensorId}`, {
    method: 'PUT',
    body: JSON.stringify(sensor),
  });
}

export async function deleteSensor(sensorId) {
  return fetchAPI(`/admin/sensors/${sensorId}`, {
    method: 'DELETE',
  });
}

// Cameras
export async function getCameras() {
  return fetchAPI('/cameras');
}

export async function createCamera(camera) {
  return fetchAPI('/admin/cameras', {
    method: 'POST',
    body: JSON.stringify(camera),
  });
}

export async function updateCamera(cameraId, camera) {
  return fetchAPI(`/admin/cameras/${cameraId}`, {
    method: 'PUT',
    body: JSON.stringify(camera),
  });
}

export async function deleteCamera(cameraId) {
  return fetchAPI(`/admin/cameras/${cameraId}`, {
    method: 'DELETE',
  });
}

// Alerts
export async function getAlerts() {
  return fetchAPI('/alerts');
}

export async function acknowledgeAlert(alertId) {
  return fetchAPI(`/admin/alerts/${alertId}/acknowledge`, {
    method: 'PUT',
  });
}

export async function deleteAlert(alertId) {
  return fetchAPI(`/admin/alerts/${alertId}`, {
    method: 'DELETE',
  });
}
