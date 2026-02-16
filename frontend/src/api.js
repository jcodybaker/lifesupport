// API service for communicating with the backend

const API_BASE_URL = '/api';

// Helper function for making requests
async function request(url, options = {}) {
  const response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`API Error: ${response.status} - ${errorText}`);
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return null;
  }

  return response.json();
}

// Device API
export const deviceAPI = {
  async list() {
    // Note: Backend may need GET /api/devices endpoint for listing all devices
    return request('/devices');
  },

  async get(deviceId) {
    return request(`/devices/${deviceId}`);
  },

  async create(device, subsystemId) {
    return request('/devices', {
      method: 'POST',
      body: JSON.stringify({ device, subsystem_id: subsystemId }),
    });
  },

  async update(deviceId, device) {
    return request(`/devices/${deviceId}`, {
      method: 'PUT',
      body: JSON.stringify(device),
    });
  },

  async delete(deviceId) {
    return request(`/devices/${deviceId}`, {
      method: 'DELETE',
    });
  },
};

// Sensor Reading API
export const sensorAPI = {
  async getReadings(params = {}) {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, value);
      }
    });
    const query = queryParams.toString();
    return request(`/sensor-readings${query ? '?' + query : ''}`);
  },

  async getLatest(sensorId) {
    return request(`/sensor-readings/${sensorId}/latest`);
  },

  async create(reading) {
    return request('/sensor-readings', {
      method: 'POST',
      body: JSON.stringify(reading),
    });
  },
};

// Actuator State API
export const actuatorAPI = {
  async getStates(params = {}) {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, value);
      }
    });
    const query = queryParams.toString();
    return request(`/actuator-states${query ? '?' + query : ''}`);
  },

  async getLatest(actuatorId) {
    return request(`/actuator-states/${actuatorId}/latest`);
  },

  async create(state) {
    return request('/actuator-states', {
      method: 'POST',
      body: JSON.stringify(state),
    });
  },
};

// Workflow API
export const workflowAPI = {
  async startDiscovery(options = {}) {
    return request('/workflows/discovery', {
      method: 'POST',
      body: JSON.stringify({ options }),
    });
  },

  async getStatus(workflowId) {
    return request(`/workflows/${workflowId}`);
  },

  async list() {
    return request('/workflows');
  },
};
