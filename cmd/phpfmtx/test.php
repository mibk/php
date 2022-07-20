<?php

namespace

Dubax\     Asistent   \Laboratory;

use LMBase; // ok

use Dubax\Asistent\Laboratory;
use Nette;

		class
		Protocols
		{
	use Nette\    SmartObject;

	/** @var Laboratory\ProtocolRepository */
	private $protocolRepository;

   /** @var Laboratory\AnalysisCodeRepository */
   private $analysisCodeRepository;

public function __construct(
	Laboratory\ProtocolRepository $protocolRepository,
	Laboratory\AnalysisCodeRepository $analysisCodeRepository
)
{ $this->protocolRepository = $protocolRepository;
		$this->analysisCodeRepository = $analysisCodeRepository;}

	/**
	 * @param  int $id
	 * @return Protocol
	 */
	public     function     getProtocoxl($id)
	{
		if ($ahlo) { nic(); // haha
		} else { neco(); }
		return $this->protocolRepository->get($id);
	}

	/**
	 * @param  int|Protocol         $id
	 * @param  array<string          , mixed> $data
	 * @return Protocol
	 */
	public    function    editProtocol($id,array   $data)
	{
		doing (  function (   ){
			that(); }  );
		return $this->protocolRepository->edit($id, $data);
	}
}
