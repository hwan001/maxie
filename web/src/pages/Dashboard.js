import React, { useState, useEffect } from 'react';
import axios from 'axios';

import Navbar from '../components/Navbar';
import Editor from '../components/Editor';

import {Title, Logo, MenuItems} from '../constants';

const ClientDataPage = () => {
  const [clientData, setClientData] = useState([]);
  const [clientId, setClientId] = useState(''); // 조회할 클라이언트 ID
  
  // useEffect(() => {
  //   if (clientId) {
  //     axios.get(`/api/client-data?client_id=${clientId}`)
  //       .then(response => {
  //         setClientData(response.data.data);
  //       })
  //       .catch(error => {
  //         console.error('Error fetching client data:', error);
  //       });
  //   }
  // }, [clientId]);

  useEffect(() => {
    axios.get(`/api/data`)
      .then(response => {
        setClientData(response.data.data);
      })
      .catch(error => {
        console.error('Error fetching client data:', error);
      });
  }, []);

    return (
        <>
            <Navbar title={Title} logo={Logo} menuItems={MenuItems["dashboard"]} />
            <Editor data="test data" />
        </>
    );
};

export default ClientDataPage;
