import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from "react-router-dom";
import api from "../components/api";
import { Shield, Home, History, Info, UserCircle, LogOut, X, LayoutDashboard } from 'lucide-react';
import { handleError } from './utils';

// The Navbar component is designed to be self-contained and reusable.
// It includes a logo, navigation links, and a user profile card.
const Navbar = () => {
    const navigate = useNavigate();

  const Logout = () => {
    localStorage.clear();
    setTimeout(() => {
      navigate("/login");
    }, 1500);
  };

  // State to manage the visibility of the profile card.
  const [showProfileCard, setShowProfileCard] = useState(false);
  const [userData, setUserData] = useState({
    username: '',
    email: '',
    attacks_count: 0
  });

  const fetchUserData = async () => {
    try {
      const res = await api.get('/api/auth/getUserInfo');
      const data = await res.data;
      if (data && data.data) {
        setUserData(data.data);
      }
    } catch (err) {
      handleError('Error fetching user data:', err);
    }
  };

  // Fetch user data when component mounts
  useEffect(() => {
    fetchUserData();
  }, []);

  // The main JSX for the Navbar.
  return (
    <nav className="fixed top-0 inset-x-0 z-20 bg-black/40 backdrop-blur-md rounded-b-xl px-6 py-4 shadow-xl">
      <div className="container mx-auto flex items-center justify-between">
        {/* Logo/App Title */}
        <a className="flex items-center text-white text-xl font-bold tracking-wider" href="#">
          <Shield className="w-8 h-8 mr-2 text-green-400" />
          VULN<span className="text-green-400">ORA</span>
        </a>

        {/* Navigation Links */}
        <div className="flex items-center space-x-6">
          <Link className="flex items-center text-gray-400 hover:text-green-400 transition-colors" to="/">
            <Home className="w-5 h-5 mr-1" /> Home
          </Link>
          <Link className="flex items-center text-gray-400 hover:text-green-400 transition-colors" to="/home">
            <LayoutDashboard className="w-5 h-5 mr-1" /> Dashboard
          </Link>
          <Link className="flex items-center text-gray-400 hover:text-green-400 transition-colors" to="/history">
            <History className="w-5 h-5 mr-1" /> History
          </Link>
          <Link className="flex items-center text-gray-400 hover:text-green-400 transition-colors" to="/#about">
            <Info className="w-5 h-5 mr-1" /> About
          </Link>
        </div>

        {/* User Profile Icon and Card */}
        <div className="relative">
          <div
            className="text-gray-400 hover:text-white transition-colors cursor-pointer"
            onClick={() => setShowProfileCard(!showProfileCard)}
          >
            <UserCircle className="w-8 h-8" />
          </div>

          {/* The profile card, conditionally rendered */}
          {showProfileCard && (
            <div className="absolute top-14 right-0 mt-2 w-64 bg-black/70 backdrop-blur-md rounded-xl shadow-2xl p-4 border border-green-400/20 text-white z-30 animate-fade-in-down">
              {/* Close Button for the profile card */}
              <button
                onClick={() => setShowProfileCard(false)}
                className="absolute top-2 right-2 p-1 rounded-full text-gray-400 hover:bg-gray-700 hover:text-white transition-colors"
              >
                <X className="w-4 h-4" />
              </button>

              <div className="flex items-center mb-4">
                <UserCircle className="w-10 h-10 text-green-400 mr-3" />
                <div>
                  <h4 className="font-bold text-lg">{userData.username || 'User'}</h4>
                  <p className="text-xs text-gray-400">{userData.email || ''}</p>
                </div>
              </div>

              <div className="text-sm border-t border-gray-700 pt-3 mb-3">
                <p className="text-gray-300">Attacks Run: <span className="text-green-400 font-semibold">
                  {userData.attacks_count || 0}
                  </span></p>
              </div>

              <button
                onClick={() => {
                  Logout();
                  setShowProfileCard(false);
                }}
                className="w-full flex items-center justify-center py-2 px-4 rounded-lg bg-red-600/50 hover:bg-red-600/70 transition-all text-white font-bold text-sm"
              >
                <LogOut className="w-4 h-4 mr-2" /> Logout
              </button>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
}

export default Navbar;
