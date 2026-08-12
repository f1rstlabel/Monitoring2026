import axios from 'axios';

const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json'
  }
});

let cachedClientIp = '';

if (typeof window !== 'undefined') {
  if (window.location.hostname && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1' && window.location.hostname !== '::1') {
    cachedClientIp = window.location.hostname;
  } else {
    fetch('https://api.ipify.org?format=json')
      .then(res => res.json())
      .then(data => {
        if (data && data.ip) {
          cachedClientIp = data.ip;
        }
      })
      .catch(() => {});
  }
}

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('gov_monitor_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    if (cachedClientIp) {
      config.headers['X-Client-IP'] = cachedClientIp;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('gov_monitor_token');
      if (window.location.pathname !== '/login') {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export default api;
