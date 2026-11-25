import Logo from '/src/assets/Home.png';

interface Props {
  onLogout: () => void;
}

export function Navbar({ onLogout }: Props) {

  return (
    <nav> 
       <a href="#Home" className='Home'><img src={Logo} alt="Logo do cin" className='LogoNav' /></a>
       <input type="text" placeholder='O que vamos ouvir?'/>
          <div className="nav_bar">
            <button className="logout-btn" onClick={onLogout}>
              Sair
            </button>
          </div>
      </nav>
  )
}
