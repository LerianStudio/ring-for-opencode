import { useState } from 'react';
import axios from 'axios';

export interface User {
  id: number;
  name: string;
}

export type Status = 'active' | 'inactive';

export function greet(name: string): string {
  return `Hello, ${name}!`;
}

export async function fetchUser(id: number): Promise<User> {
  const response = await axios.get(`/users/${id}`);
  return response.data;
}

export const formatName = (name: string): string => {
  return name.trim().toUpperCase();
};

export class UserService {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async getUser(id: number): Promise<User> {
    const response = await axios.get(`${this.baseUrl}/users/${id}`);
    return response.data;
  }
}
