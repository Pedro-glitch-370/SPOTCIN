import Logo from '/src/assets/Home.png';

function Navbar() {

  return (
    <nav> 
       <a href="#Home" className='Home'><img src={Logo} alt="Logo do cin" className='LogoNav' /></a>
       <input type="text" placeholder='O que vamos ouvir?'/>
          <div className="nav_bar">
                <div className="log_cad">
                    <a href="#login" className='Login'>Login</a>
                    <a href="#cadastro" className="cadastro">Cadastre-se</a>
                </div>
          </div>
      </nav>
  )
}

export default Navbar
