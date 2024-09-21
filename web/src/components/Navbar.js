import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Button } from './Button';

import Cookies from 'js-cookie'; 
import '../styles/Navbar.css';

function Navbar({ menuItems }) {
    const [click, setClick] = useState(false);
    const handleClick = () => setClick(!click);
    const closeMobileMenu = () => setClick(false);

    const [isMenuOpen, setIsMenuOpen] = useState(false);
    const toggleMenu = () => {
      setIsMenuOpen(!isMenuOpen);
    };
    
    const [session, setSession] = useState(false);
    const getJwt = () => {
        /* 백엔드 미들웨어로 인증 받아서 처리 */
        const jwt = Cookies.get('jwt');
        if (jwt) {
            console.log(jwt);
            setSession(true);
        } else {
            setSession(false);
        }
    }
    useEffect(() => { getJwt(); }, []);

    return (
        <>
            <nav className='navbar'>
                <div className='navbar-container'>
                    <Link to='/' className='navbar-logo' onClick={closeMobileMenu}>
                        LOGO
                        <i className='fab fa-typo3' />
                    </Link>
                    <div className='menu-icon' onClick={handleClick}>
                        <i className={click ? 'fas fa-times' : 'fas fa-bars'} />
                    </div>
                    <ul className={click ? 'nav-menu active' : 'nav-menu'}>
                        {menuItems.map((item, index) => (
                            <li className='nav-item' key={index}>
                                <Link 
                                      to={item.path} 
                                      className='nav-links' 
                                      onClick={closeMobileMenu}
                                    >
                                      {item.label}
                                </Link>
                                {item.submenu ? (
                                    <div className='dropdown-content'>
                                      {item.submenu.map((submenuItem, subIndex) => (
                                        <a href={submenuItem.path} key={index + subIndex}>
                                          <div>
                                              <img src={submenuItem.img}/>
                                              <div>{submenuItem.label}</div>
                                          </div>
                                        </a>
                                      ))}
                                    </div>
                                ) : (
                                  <></>
                                )}
                            </li>
                        ))}

                        {/* 모바일 메뉴 토글 버튼 */}
                        <div className="menu-toggle" onClick={toggleMenu} > 
                            {/* <span className="menu-icon"></span> */}
                            <i className="fas fa-bars"></i>
                        </div>
                    </ul>
                    {!session && <Button linkTo='/login' buttonStyle='btn--outline'>Log in</Button>}
                    {!session && <Button linkTo='/signup' buttonStyle='btn--outline'>Sign up</Button>}

                     {/* 모바일 메뉴 */}
                    {/* !button && <MobileMenu />*/}
                </div>
            </nav>
        </>
    );
}

export default Navbar;


const MobileMenu = () => {
  return (
    <div className="mobile-menu">
      <ul className="mobile-nav-links">
        <li><Link to="/download">Download</Link></li>
        <li><Link to="/docs">Docs</Link></li>
        <li><Link to="/blog">Blog</Link></li>
      </ul>
      <div className="mobile-auth-buttons">
        { /* <Link to="/login" className="btn btn--outline">Login</Link>
        <Link to="/signup" className="btn btn--primary">Sign Up</Link> */}
        <Button linkTo='/login' buttonStyle='mobile-btn--outline'>Log in</Button>
        <Button linkTo='/signup' buttonStyle='mobile-btn--outline'>Sign up</Button>
      </div>
    </div>
  );
};