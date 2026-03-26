import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class BackendApiService {
  private apiUrl = `http://${window.location.hostname}:8002`; // Use environment variable in production

  constructor(private http: HttpClient) {}

  setUserProfile(profile: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/user-profile`, profile);
  }

  uploadPortfolio(file: File): Observable<any> {
    const formData = new FormData();
    formData.append('file', file);
    return this.http.post(`${this.apiUrl}/upload-portfolio`, formData);
  }

  getSuggestions(): Observable<any> {
    return this.http.get(`${this.apiUrl}/suggest`);
  }

  getExplanation(): Observable<any> {
    return this.http.get(`${this.apiUrl}/explain`);
  }
}
