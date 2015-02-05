if [[ ! -s "$HOME/.bash_profile" && -s "$HOME/.profile" ]] ; then
  profile_file="$HOME/.profile"
else
  profile_file="$HOME/.bash_profile"
fi

if ! grep -q 'bash_autocomplete' "${profile_file}" ; then
  echo "PROG=cx source /usr/local/cloud66/bash_autocomplete" >> "${profile_file}"
fi
