import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useGoogleLogin } from '@react-oauth/google';
import '../styles/Login.css';
import IMG_PROFILE from '../images/profile.png';

import Navbar from "../components/Navbar.js";

import { BASE_URL } from '../constants';

function Profile() {
  const [profile, setProfile] = useState(null);
  const [error, setError] = useState(null);

  // Google Login 기능 설정
  const login = useGoogleLogin({
    flow: 'auth-code',
    onSuccess: async (codeResponse) => {
      try {
        const response = await axios.post(`${BASE_URL}/auth/google`, {
          code: codeResponse.code,
        }, { withCredentials: true });

        console.log(response.data);
        if (response.data.message === 'Logged in') {
          setProfile(response.data.profile);
        } else {
          console.error('Failed to set JWT token in cookie');
        }
      } catch (error) {
        console.error('Error during login:', error);
        setError(error);
      }
    },
    onError: errorResponse => console.log(errorResponse),
  });

  // 프로필 정보를 가져오는 함수
  const fetchProfile = async () => {
    try {
      const response = await axios.get(`${BASE_URL}/protected/profile`, { withCredentials: true });
      setProfile(response.data);
    } catch (err) {
      if (err.response && err.response.status === 401) {
        console.log('User is not authenticated');
        setError(null);
      } else {
        console.error('Failed to fetch profile', err);
        setError(err);
      }
    }
  };

  // 로그아웃 함수
  const logout = () => {
    fetch(`${BASE_URL}/protected/logout`, {
      method: 'POST',
      credentials: 'include',
    })
      .then(response => response.json())
      .then(data => {
        console.log(data);
        setProfile(null);
      })
      .catch(error => console.error('Error:', error));
  };

  // 특정 엔드포인트로 리다이렉트
  const handleRedirect = () => {
    window.location.href = `${BASE_URL}/protected`;
  };

  // 컴포넌트가 마운트될 때 프로필 정보를 가져옴
  useEffect(() => {
    fetchProfile();
  }, []);

  const MenuItems = [
    {label: "Home", path: "/"},
    { 
        label: "Downloads", 
        submenu: [
          { img: "", label: "Windows", path: "/api/download/client/windows" },
          { img: "", label: "linux", path: "/api/download/client/linux" },
          { img: "", label: "mac", path: "/api/download/client/mac" }
        ]
    }
  ];


  return (
    <>
    <Navbar menuItems={MenuItems} />
    <div className='profile-container'>
      {profile ? (
          <div className="profile-details">
            <img src={profile.picture} alt="profile" className='profile-avathar' />
            <div className="profile-content">
              <h3>{profile.name}</h3>
              <h5>{profile.email}</h5>
            </div>
            <div>
              <button className='profile-button' onClick={logout}>Logout</button>
            </div>
            <button className='profile-button' onClick={handleRedirect}>Go to Specific Endpoint</button>
          </div>
        ) : (
          <div className="landing-container">
            <div className="profile-details">
              <div className="landing-icon">
                <img src={IMG_PROFILE} alt="empty profile" className='profile-avathar' />
              </div>
              <h4>Sign in to create profile!</h4>
              <div>
                <button className='profile-button' onClick={login}>Sign in with Google</button>
              </div>
              <div>
                <button className='profile-button' onClick={handleRedirect}>Go to Specific Endpoint</button>
              </div>
            </div>
          </div>
        )
      }
    </div>
    </>
  );
}

export default Profile;
