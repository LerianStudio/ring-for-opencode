import { useState, useEffect } from 'react';
import { api } from './api';

export interface User {
  id: number;
  name: string;
  email: string;  // Added field
  isActive: boolean;  // Added field
}

export type Status = 'active' | 'inactive' | 'pending';  // Added 'pending'

export interface Config {  // New interface
  debug: boolean;
  timeout: number;
}

export function greet(name: string, greeting?: string): string {  // Added parameter
  return `${greeting || 'Hello'}, ${name}!`;
}

export async function fetchUser(id: number, options?: Config): Promise<User> {  // Added parameter
  const response = await api.get(`/users/${id}`, options);
  return response.data;
}

// formatName removed

export const validateEmail = (email: string): boolean => {  // New function
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
};

export class UserService {
  private baseUrl: string;
  private config: Config;  // Added field

  constructor(baseUrl: string, config: Config) {  // Changed signature
    this.baseUrl = baseUrl;
    this.config = config;
  }

  async getUser(id: number): Promise<User> {
    const response = await api.get(`${this.baseUrl}/users/${id}`);
    return response.data;
  }

  async updateUser(id: number, data: Partial<User>): Promise<User> {  // New method
    const response = await api.put(`${this.baseUrl}/users/${id}`, data);
    return response.data;
  }
}
