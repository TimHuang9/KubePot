import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button, message, Dropdown, Avatar } from 'antd';
import { LogoutOutlined, UserOutlined } from '@ant-design/icons';
import dashboardApi from '@/api/dashboard-api';

const Header = () => {
  const [username, setUsername] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    const storedUsername = localStorage.getItem('username');
    if (storedUsername) {
      setUsername(storedUsername);
    }
  }, []);

  const handleLogout = async () => {
    try {
      await dashboardApi.logout();
      localStorage.removeItem('is_login');
      localStorage.removeItem('username');
      message.success('已登出');
      navigate('/login');
    } catch (error) {
      console.error('登出失败:', error);
      localStorage.removeItem('is_login');
      localStorage.removeItem('username');
      navigate('/login');
    }
  };

  const items = [
    {
      key: 'logout',
      label: '登出',
      icon: <LogoutOutlined />,
      onClick: handleLogout,
    },
  ];

  return (
    <header className="bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg">
      <div className="container mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          <Link to="/" className="flex items-center space-x-2">
            <div className="w-10 h-10 bg-white rounded-lg flex items-center justify-center">
              <span className="text-blue-600 font-bold text-xl">K</span>
            </div>
            <span className="text-2xl font-bold">KubePot</span>
          </Link>
          
          <nav className="flex items-center space-x-6">
            <Link to="/" className="hover:text-blue-200 transition-colors">
              仪表盘
            </Link>
            <div className="flex items-center space-x-2">
              <Dropdown menu={{ items }} placement="bottomRight">
                <div className="flex items-center space-x-2 cursor-pointer hover:bg-white/10 px-3 py-1 rounded-lg">
                  <Avatar icon={<UserOutlined />} size="small" className="bg-white/20" />
                  <span>{username || '用户'}</span>
                </div>
              </Dropdown>
            </div>
          </nav>
        </div>
      </div>
    </header>
  );
};

export default Header;
