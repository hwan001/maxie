import React, { useState, useEffect } from 'react';
import axios from 'axios';

import Navbar from '../components/Navbar';
import Editor from '../components/Editor';

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

  const MenuItems = [
    { i:"fa-solid fa-house", label: "Home", path:"/"},
    { label: "Blog", path: "https://hwan001.co.kr"},
    { label: "GitHub", path: "https://github.com/hwan001"},
  ];

    return (
        <>
            <Navbar menuItems={MenuItems} />
            {/* <div>
              <h1>Client Data</h1>
              <input
                type="text"
                value={clientId}
                onChange={(e) => setClientId(e.target.value)}
                placeholder="Enter Client ID"
              />
              <button onClick={() => setClientId(clientId)}>Fetch Data</button>
              <table>
                <thead>
                  <tr>
                    <th>Client ID</th>
                    <th>Network Interfaces</th>
                    <th>System Info</th>
                    <th>Active Ports</th>
                    <th>Timestamp</th>
                  </tr>
                </thead>
                <tbody>
                  {clientData.map((data, index) => (
                    <tr key={index}>
                      <td>{data.ClientID}</td>
                      <td>{data.NetworkInterfaces}</td>
                      <td>{data.SystemInfo}</td>
                      <td>{data.ActivePorts}</td>
                      <td>{data.CreatedAt}</td>
                    </tr>
                  ))}
                </tbody>
              </table> 
            </div> */}
            <div><Editor data="test data"></Editor></div>
        </>
    );
};

export default ClientDataPage;
